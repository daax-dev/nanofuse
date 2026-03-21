# GOAL: 
to build a custom build a firecracker image, that could work with something like slicer https://docs.slicervm.com/getting-started/walkthrough/. and be pulled down and run exactly like slicer (this is a complete rebuild for learning purposes) so lets document the phased planned approach)

# KEY INFORMATION

Use ubuntu 24.04, for a tiny base image. 
want it to be built via github actions (make where needed)
assume pushing to a private GH artifact registry for install (like slicer)
this is for a foundation for building isloated microvms for the 2 sides of trigger.dev (web + supervisor/worker) see here https://trigger.dev/docs/self-hosting/overview

# componets

1. CLI - to invoke on target machine - to pull down firecracker kernel + root FS (and boot)
2. API - systemd service that runs and can stop start resume the images.
3. the MicroVM Image - that gets pulled down and used.