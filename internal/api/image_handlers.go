package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/daax-dev/nanofuse/internal/types"
	"github.com/daax-dev/nanofuse/internal/vmm"
	"github.com/google/uuid"
)

// handleImages handles image list operations
func (s *Server) handleImages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		types.WriteError(w, http.StatusMethodNotAllowed, types.ErrInvalidRequest, "Method not allowed", nil)
		return
	}

	images, err := s.db.ListImages()
	if err != nil {
		s.logger.Error("Failed to list images: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to list images", nil)
		return
	}
	if provider, ok := s.runtimeManager.(vmm.ImageProvider); ok {
		runtimeImages, runtimeErr := provider.ListImages()
		if runtimeErr != nil {
			s.logger.Warn("Failed to list runtime images: %v", runtimeErr)
		} else {
			images = mergeRuntimeImages(images, runtimeImages)
		}
	}

	response := types.ListImagesResponse{
		Images: make([]types.Image, 0, len(images)),
		Total:  len(images),
	}

	for _, img := range images {
		response.Images = append(response.Images, *img)
	}

	writeJSON(w, http.StatusOK, response)
}

func mergeRuntimeImages(dbImages, runtimeImages []*types.Image) []*types.Image {
	seen := make(map[string]struct{}, len(dbImages)+len(runtimeImages))
	merged := make([]*types.Image, 0, len(dbImages)+len(runtimeImages))
	for _, image := range dbImages {
		key := image.Digest
		if key == "" && len(image.Tags) > 0 {
			key = image.Tags[0]
		}
		if key != "" {
			seen[key] = struct{}{}
		}
		merged = append(merged, image)
	}
	for _, image := range runtimeImages {
		key := image.Digest
		if key == "" && len(image.Tags) > 0 {
			key = image.Tags[0]
		}
		if key != "" {
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
		}
		merged = append(merged, image)
	}
	return merged
}

// handleGetImage gets a specific image
func (s *Server) handleGetImage(w http.ResponseWriter, r *http.Request, digest string) {
	image, err := s.db.GetImage(digest)
	if err != nil {
		s.logger.Error("Failed to get image %s: %v", digest, err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to get image", nil)
		return
	}

	if image == nil {
		types.WriteError(w, http.StatusNotFound, types.ErrImageNotFound,
			fmt.Sprintf("Image with digest '%s' not found", digest),
			map[string]interface{}{"digest": digest})
		return
	}

	writeJSON(w, http.StatusOK, image)
}

// handleDeleteImage deletes an image
func (s *Server) handleDeleteImage(w http.ResponseWriter, r *http.Request, digest string) {
	s.logger.Debug("Delete image request for: %s", digest)

	// Check if any VMs use this image
	vms, err := s.db.ListVMs("")
	if err != nil {
		s.logger.Error("Failed to list VMs while checking image usage: %v", err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to check image usage", nil)
		return
	}

	var usingVMs []string
	for _, vm := range vms {
		if vm.ImageDigest == digest {
			usingVMs = append(usingVMs, vm.Name)
		}
	}

	if len(usingVMs) > 0 {
		s.logger.Warn("Cannot delete image %s: in use by VMs: %v", digest, usingVMs)
		types.WriteError(w, http.StatusConflict, types.ErrResourceInUse,
			fmt.Sprintf("Image is in use by %d VMs", len(usingVMs)),
			map[string]interface{}{
				"vms":   usingVMs,
				"count": len(usingVMs),
			})
		return
	}

	// Delete from database
	if err := s.db.DeleteImage(digest); err != nil {
		s.logger.Error("Failed to delete image %s from database: %v", digest, err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to delete image", nil)
		return
	}

	// TODO: Delete image files from disk

	s.logger.Info("Deleted image: %s", digest)
	w.WriteHeader(http.StatusNoContent)
}

// handleImagePull handles image pull requests
func (s *Server) handleImagePull(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		types.WriteError(w, http.StatusMethodNotAllowed, types.ErrInvalidRequest, "Method not allowed", nil)
		return
	}

	var req types.PullImageRequest
	if err := readJSON(r, &req); err != nil {
		s.logger.Warn("Invalid pull request body: %v", err)
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "Invalid request body", nil)
		return
	}

	if req.ImageRef == "" {
		s.logger.Warn("Pull request missing image_ref")
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "image_ref is required", nil)
		return
	}

	s.logger.Info("Received pull request for image: %s", req.ImageRef)

	// Create pull job
	jobID := "job-" + uuid.New().String()
	job := &types.ImagePullJob{
		ID:        jobID,
		ImageRef:  req.ImageRef,
		State:     types.PullJobPending,
		CreatedAt: time.Now(),
	}

	if err := s.db.CreatePullJob(job); err != nil {
		s.logger.Error("Failed to create pull job for %s: %v", req.ImageRef, err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to create pull job", nil)
		return
	}

	s.logger.Info("Created pull job %s for image: %s", jobID[:8], req.ImageRef)

	// Start async pull
	go s.executePullJob(jobID, req.ImageRef)

	// Return job info
	response := types.PullImageResponse{
		JobID:     jobID,
		ImageRef:  req.ImageRef,
		State:     types.PullJobPending,
		StatusURL: fmt.Sprintf("/images/jobs/%s", jobID),
	}

	writeJSON(w, http.StatusAccepted, response)
}

// executePullJob executes an image pull job
func (s *Server) executePullJob(jobID, imageRef string) {
	startTime := time.Now()
	s.logger.Info("[job:%s] Starting pull job for image: %s", jobID[:8], imageRef)

	// Update job state to in_progress
	job, err := s.db.GetPullJob(jobID)
	if err != nil {
		s.logger.Error("[job:%s] Failed to get pull job from database: %v", jobID[:8], err)
		return
	}

	job.State = types.PullJobInProgress
	if err := s.db.UpdatePullJob(job); err != nil {
		s.logger.Error("[job:%s] Failed to update pull job state to in_progress: %v", jobID[:8], err)
		return
	}
	s.logger.Debug("[job:%s] Job state updated to in_progress", jobID[:8])

	// Create progress channel with done signal
	progressChan := make(chan *types.PullProgress, 10)
	progressDone := make(chan struct{})
	go func() {
		defer close(progressDone)
		for progress := range progressChan {
			job.Progress = progress
			if err := s.db.UpdatePullJob(job); err != nil {
				s.logger.Warn("[job:%s] Failed to update pull progress: %v", jobID[:8], err)
			} else {
				s.logger.Debug("[job:%s] Progress: %d/%d bytes (%d%%)",
					jobID[:8], progress.CurrentBytes, progress.TotalBytes, progress.Percentage)
			}
		}
		s.logger.Debug("[job:%s] Progress channel closed", jobID[:8])
	}()

	if provider, ok := s.runtimeManager.(vmm.ImageProvider); ok {
		close(progressChan)
		<-progressDone

		image, err := provider.ResolveImage(imageRef)
		if err != nil {
			s.logger.Error("[job:%s] Runtime pull failed after %v: %v", jobID[:8], time.Since(startTime), err)
			job.State = types.PullJobFailed
			errMsg := err.Error()
			job.Error = &errMsg
			now := time.Now()
			job.CompletedAt = &now
			if updateErr := s.db.UpdatePullJob(job); updateErr != nil {
				s.logger.Error("[job:%s] Failed to update job state to failed: %v", jobID[:8], updateErr)
			}
			return
		}
		if err := s.db.UpsertImage(image); err != nil {
			s.logger.Error("[job:%s] Failed to save runtime image to database: %v", jobID[:8], err)
			job.State = types.PullJobFailed
			errMsg := fmt.Sprintf("Failed to save image: %v", err)
			job.Error = &errMsg
			now := time.Now()
			job.CompletedAt = &now
			if updateErr := s.db.UpdatePullJob(job); updateErr != nil {
				s.logger.Error("[job:%s] Failed to update job state to failed: %v", jobID[:8], updateErr)
			}
			return
		}
		job.State = types.PullJobCompleted
		job.ResultDigest = &image.Digest
		now := time.Now()
		job.CompletedAt = &now
		if err := s.db.UpdatePullJob(job); err != nil {
			s.logger.Error("[job:%s] Failed to update job state to completed: %v", jobID[:8], err)
		}
		s.logger.Info("[job:%s] Runtime pull job completed in %v: %s -> %s", jobID[:8], time.Since(startTime), imageRef, image.Digest)
		return
	}

	// Pull image with a proper timeout context
	// Use the configured pull timeout from the registry client
	pullTimeout := time.Duration(s.config.Registry.PullTimeoutSecs) * time.Second
	if pullTimeout == 0 {
		pullTimeout = 10 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), pullTimeout)
	defer cancel()

	s.logger.Info("[job:%s] Calling registry client PullImage (timeout: %v)", jobID[:8], pullTimeout)
	image, err := s.registryClient.PullImage(ctx, imageRef, progressChan)
	close(progressChan)

	// Wait for progress goroutine to finish before updating job state
	// This prevents race condition where progress update overwrites final state
	<-progressDone

	if err != nil {
		s.logger.Error("[job:%s] Pull failed after %v: %v", jobID[:8], time.Since(startTime), err)
		job.State = types.PullJobFailed
		errMsg := err.Error()
		job.Error = &errMsg
		now := time.Now()
		job.CompletedAt = &now
		if updateErr := s.db.UpdatePullJob(job); updateErr != nil {
			s.logger.Error("[job:%s] Failed to update job state to failed: %v", jobID[:8], updateErr)
		}
		return
	}

	s.logger.Info("[job:%s] Registry pull completed, saving image to database...", jobID[:8])

	// Save image to database
	if err := s.db.CreateImage(image); err != nil {
		s.logger.Error("[job:%s] Failed to save image to database: %v", jobID[:8], err)
		job.State = types.PullJobFailed
		errMsg := fmt.Sprintf("Failed to save image: %v", err)
		job.Error = &errMsg
		now := time.Now()
		job.CompletedAt = &now
		if updateErr := s.db.UpdatePullJob(job); updateErr != nil {
			s.logger.Error("[job:%s] Failed to update job state to failed: %v", jobID[:8], updateErr)
		}
		return
	}
	s.logger.Debug("[job:%s] Image saved to database", jobID[:8])

	// Update job as completed
	job.State = types.PullJobCompleted
	job.ResultDigest = &image.Digest
	now := time.Now()
	job.CompletedAt = &now
	if err := s.db.UpdatePullJob(job); err != nil {
		s.logger.Error("[job:%s] Failed to update job state to completed: %v", jobID[:8], err)
	}

	totalDuration := time.Since(startTime)
	s.logger.Info("[job:%s] Pull job completed in %v: %s -> %s", jobID[:8], totalDuration, imageRef, image.Digest)
}

// ============================================================================
// Go 1.22+ Path Parameter Wrappers
// ============================================================================

// handleImageJobByPath handles GET /images/jobs/{id} using path parameters
func (s *Server) handleImageJobByPath(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")
	if jobID == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "Job ID is required", nil)
		return
	}

	job, err := s.db.GetPullJob(jobID)
	if err != nil {
		s.logger.Error("Failed to get pull job %s: %v", jobID, err)
		types.WriteError(w, http.StatusInternalServerError, types.ErrInternalError, "Failed to get pull job", nil)
		return
	}

	if job == nil {
		types.WriteError(w, http.StatusNotFound, types.ErrImageNotFound,
			fmt.Sprintf("Pull job '%s' not found", jobID),
			map[string]interface{}{"job_id": jobID})
		return
	}

	writeJSON(w, http.StatusOK, job)
}

// handleGetImageByPath handles GET /images/{digest} using path parameters
func (s *Server) handleGetImageByPath(w http.ResponseWriter, r *http.Request) {
	digest := r.PathValue("digest")
	if digest == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "Image digest is required", nil)
		return
	}
	s.handleGetImage(w, r, digest)
}

// handleDeleteImageByPath handles DELETE /images/{digest} using path parameters
func (s *Server) handleDeleteImageByPath(w http.ResponseWriter, r *http.Request) {
	digest := r.PathValue("digest")
	if digest == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest, "Image digest is required", nil)
		return
	}
	s.handleDeleteImage(w, r, digest)
}
