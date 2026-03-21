# Kotlin Error Handling for Mobile Development

*Comprehensive guide to error handling patterns, strategies, and best practices in Kotlin mobile applications*

## Table of Contents

1. [Core Error Handling Mechanisms](#core-error-handling-mechanisms)
2. [Kotlin-Specific Error Handling](#kotlin-specific-error-handling)
3. [Sealed Classes for Error States](#sealed-classes-for-error-states)
4. [Android-Specific Error Handling](#android-specific-error-handling)
5. [Coroutines Error Handling](#coroutines-error-handling)
6. [Reactive Error Handling](#reactive-error-handling)
7. [Functional Error Handling](#functional-error-handling)
8. [Error Recovery Strategies](#error-recovery-strategies)
9. [Testing Error Handling](#testing-error-handling)
10. [Error Logging and Monitoring](#error-logging-and-monitoring)

---

## Core Error Handling Mechanisms

### Exception Hierarchy and Types

Kotlin inherits Java's exception hierarchy but with important distinctions:

```kotlin
// All exceptions in Kotlin are unchecked (runtime exceptions)
class NetworkException(message: String, cause: Throwable? = null) : Exception(message, cause)
class DatabaseException(message: String) : RuntimeException(message)
class ValidationException(field: String, value: Any?) : IllegalArgumentException("Invalid $field: $value")

// Custom mobile-specific exceptions
class LocationPermissionException : SecurityException("Location permission denied")
class CameraUnavailableException : RuntimeException("Camera not available")
class InsufficientStorageException : RuntimeException("Insufficient storage space")
```

### Try-Catch-Finally Blocks

```kotlin
class UserRepository(
    private val apiService: ApiService,
    private val localStorage: UserLocalStorage,
    private val crashlytics: FirebaseCrashlytics
) {
    suspend fun getUserProfile(userId: String): User {
        return try {
            // Primary operation
            val user = apiService.getUser(userId)
            localStorage.cacheUser(user)
            user
        } catch (e: NetworkException) {
            // Specific network error handling
            crashlytics.recordException(e)
            // Fallback to cached data
            localStorage.getCachedUser(userId)
                ?: throw UserNotFoundException("User $userId not found locally", e)
        } catch (e: SecurityException) {
            // Handle authentication/authorization errors
            crashlytics.log("Authentication failed for user: $userId")
            throw e
        } catch (e: Exception) {
            // Generic error handling
            crashlytics.recordException(e)
            throw UserServiceException("Failed to retrieve user profile", e)
        } finally {
            // Always executed - cleanup resources
            cleanupTempFiles()
            logRequestMetrics(userId)
        }
    }
}
```

### Custom Exception Design for Mobile

```kotlin
// Base exception for all app-specific errors
abstract class AppException(
    message: String,
    cause: Throwable? = null,
    val errorCode: String? = null,
    val userMessage: String? = null
) : Exception(message, cause)

// Network-related exceptions
sealed class NetworkException(
    message: String,
    cause: Throwable? = null,
    errorCode: String? = null,
    userMessage: String? = null
) : AppException(message, cause, errorCode, userMessage) {

    class NoConnectionException : NetworkException(
        message = "No network connection available",
        userMessage = "Please check your internet connection and try again"
    )

    class TimeoutException : NetworkException(
        message = "Network request timed out",
        userMessage = "Request timed out. Please try again"
    )

    class ServerException(val httpCode: Int, serverMessage: String) : NetworkException(
        message = "Server error: $httpCode - $serverMessage",
        userMessage = "Something went wrong. Please try again later"
    )
}

// Usage in mobile app
class ApiService {
    suspend fun makeRequest(url: String): String {
        return try {
            httpClient.get(url)
        } catch (e: ConnectException) {
            throw NetworkException.NoConnectionException()
        } catch (e: SocketTimeoutException) {
            throw NetworkException.TimeoutException()
        } catch (e: HttpException) {
            throw NetworkException.ServerException(e.code(), e.message())
        }
    }
}
```

### Precondition Functions

```kotlin
// Kotlin provides built-in precondition functions
class BankAccount(initialBalance: Double) {
    init {
        require(initialBalance >= 0) { "Initial balance must be non-negative: $initialBalance" }
    }

    private var _balance: Double = initialBalance
    val balance: Double get() = _balance

    fun withdraw(amount: Double) {
        require(amount > 0) { "Withdrawal amount must be positive: $amount" }
        check(_balance >= amount) { "Insufficient balance: $_balance < $amount" }

        _balance -= amount
    }

    fun deposit(amount: Double) {
        require(amount > 0) { "Deposit amount must be positive: $amount" }
        _balance += amount
    }
}

// Mobile-specific preconditions
class LocationService(private val context: Context) {
    fun requestLocation(): Location {
        requireNotNull(context) { "Context cannot be null" }

        val locationManager = context.getSystemService(Context.LOCATION_SERVICE) as LocationManager
        require(locationManager.isProviderEnabled(LocationManager.GPS_PROVIDER)) {
            "GPS provider is not enabled"
        }

        checkNotNull(getLastKnownLocation()) { "No recent location available" }
        return getLastKnownLocation()!!
    }
}
```

---

## Kotlin-Specific Error Handling

### Null Safety as Error Prevention

```kotlin
class UserProfileActivity : AppCompatActivity() {
    private var binding: ActivityUserProfileBinding? = null

    // Null safety prevents runtime errors
    private fun displayUser(user: User?) {
        // Safe call - method only called if user is not null
        user?.name?.let { name ->
            binding?.nameTextView?.text = name
        }

        // Elvis operator provides fallback
        val displayName = user?.name ?: "Unknown User"
        val email = user?.email ?: "No email provided"

        // Safe property access with let
        user?.profileImage?.let { imageUrl ->
            Glide.with(this)
                .load(imageUrl)
                .into(binding?.profileImageView ?: return)
        }
    }

    override fun onDestroy() {
        super.onDestroy()
        binding = null // Prevent memory leaks
    }
}
```

### Result Type for Success/Failure

```kotlin
// Kotlin's built-in Result type
class ApiRepository {
    suspend fun fetchUserData(userId: String): Result<User> {
        return Result.runCatching {
            val response = apiService.getUser(userId)
            if (response.isSuccessful) {
                response.body() ?: throw IllegalStateException("Empty response body")
            } else {
                throw HttpException(response.code(), response.message())
            }
        }
    }

    // Transform and chain Results
    suspend fun getUserDisplayName(userId: String): Result<String> {
        return fetchUserData(userId)
            .map { user -> "${user.firstName} ${user.lastName}" }
            .recover { exception ->
                when (exception) {
                    is HttpException -> "Anonymous User"
                    else -> throw exception
                }
            }
    }
}

// Usage in ViewModel
class UserViewModel(private val repository: ApiRepository) : ViewModel() {
    private val _uiState = MutableLiveData<UiState<User>>()
    val uiState: LiveData<UiState<User>> = _uiState

    fun loadUser(userId: String) {
        viewModelScope.launch {
            _uiState.value = UiState.Loading

            repository.fetchUserData(userId)
                .onSuccess { user ->
                    _uiState.value = UiState.Success(user)
                }
                .onFailure { exception ->
                    _uiState.value = UiState.Error(
                        message = exception.message ?: "Unknown error occurred",
                        exception = exception
                    )
                }
        }
    }
}
```

### Safe Operations and Elvis Operator

```kotlin
class MessageUtils {
    // Safe operations prevent crashes
    fun formatMessage(user: User?, message: String?): String {
        return buildString {
            // Safe property access
            val username = user?.name?.takeIf { it.isNotBlank() } ?: "Anonymous"
            append("[$username]: ")

            // Safe message processing
            val content = message?.trim()?.takeIf { it.isNotEmpty() } ?: "No content"
            append(content)

            // Safe timestamp formatting
            user?.lastActive?.let { timestamp ->
                append(" (${formatTimestamp(timestamp)})")
            }
        }
    }

    private fun formatTimestamp(timestamp: Long): String {
        return try {
            SimpleDateFormat("HH:mm", Locale.getDefault()).format(Date(timestamp))
        } catch (e: Exception) {
            "Unknown time"
        }
    }
}

// Android-specific safe operations
class ViewUtils {
    fun updateTextViewSafely(textView: TextView?, text: String?) {
        textView?.apply {
            this.text = text?.takeIf { it.isNotBlank() } ?: context.getString(R.string.default_text)
        }
    }

    fun loadImageSafely(imageView: ImageView?, imageUrl: String?) {
        imageView?.let { view ->
            imageUrl?.takeIf { it.isNotBlank() }?.let { url ->
                Picasso.get()
                    .load(url)
                    .error(R.drawable.placeholder_error)
                    .placeholder(R.drawable.placeholder_loading)
                    .into(view)
            } ?: run {
                view.setImageResource(R.drawable.placeholder_default)
            }
        }
    }
}
```

---

## Sealed Classes for Error States

### UI State Management

```kotlin
// Comprehensive UI state representation
sealed class UiState<out T> {
    object Loading : UiState<Nothing>()
    data class Success<T>(val data: T) : UiState<T>()
    data class Error(
        val message: String,
        val exception: Throwable? = null,
        val errorCode: String? = null,
        val isRetryable: Boolean = true
    ) : UiState<Nothing>()

    object Empty : UiState<Nothing>()
}

// Specialized states for different scenarios
sealed class AuthState {
    object Unauthenticated : AuthState()
    object Authenticating : AuthState()
    data class Authenticated(val user: User, val token: String) : AuthState()
    data class AuthenticationFailed(val error: String, val canRetry: Boolean = true) : AuthState()
    object SessionExpired : AuthState()
}

sealed class NetworkState {
    object Available : NetworkState()
    object Unavailable : NetworkState()
    data class Limited(val type: NetworkCapability) : NetworkState()
}

// Usage in ViewModel with sealed classes
class HomeViewModel(
    private val userRepository: UserRepository,
    private val networkManager: NetworkManager
) : ViewModel() {

    private val _homeState = MutableLiveData<HomeUiState>()
    val homeState: LiveData<HomeUiState> = _homeState

    sealed class HomeUiState {
        object Loading : HomeUiState()
        data class Success(
            val user: User,
            val notifications: List<Notification>,
            val hasNewContent: Boolean
        ) : HomeUiState()
        data class Error(val type: ErrorType, val message: String) : HomeUiState() {
            enum class ErrorType { NETWORK, AUTH, SERVER, UNKNOWN }
        }
        object NetworkUnavailable : HomeUiState()
    }

    fun loadHomeData() {
        viewModelScope.launch {
            _homeState.value = HomeUiState.Loading

            when (val networkState = networkManager.getNetworkState()) {
                is NetworkState.Unavailable -> {
                    _homeState.value = HomeUiState.NetworkUnavailable
                    return@launch
                }
                else -> {
                    // Proceed with network request
                }
            }

            try {
                val user = userRepository.getCurrentUser()
                val notifications = userRepository.getNotifications()

                _homeState.value = HomeUiState.Success(
                    user = user,
                    notifications = notifications,
                    hasNewContent = notifications.any { !it.isRead }
                )
            } catch (e: Exception) {
                val errorType = when (e) {
                    is NetworkException -> HomeUiState.ErrorType.NETWORK
                    is AuthenticationException -> HomeUiState.ErrorType.AUTH
                    is ServerException -> HomeUiState.ErrorType.SERVER
                    else -> HomeUiState.ErrorType.UNKNOWN
                }

                _homeState.value = HomeUiState.Error(
                    type = errorType,
                    message = e.message ?: "An error occurred"
                )
            }
        }
    }
}
```

### Exhaustive Error Handling with When

```kotlin
// Exhaustive error handling ensures all states are covered
class HomeFragment : Fragment() {
    private lateinit var binding: FragmentHomeBinding

    private fun observeHomeState() {
        viewModel.homeState.observe(viewLifecycleOwner) { state ->
            when (state) {
                is HomeViewModel.HomeUiState.Loading -> {
                    binding.progressBar.isVisible = true
                    binding.contentLayout.isVisible = false
                    binding.errorLayout.isVisible = false
                }

                is HomeViewModel.HomeUiState.Success -> {
                    binding.progressBar.isVisible = false
                    binding.contentLayout.isVisible = true
                    binding.errorLayout.isVisible = false

                    displayUserInfo(state.user)
                    displayNotifications(state.notifications)

                    if (state.hasNewContent) {
                        showNewContentIndicator()
                    }
                }

                is HomeViewModel.HomeUiState.Error -> {
                    binding.progressBar.isVisible = false
                    binding.contentLayout.isVisible = false
                    binding.errorLayout.isVisible = true

                    // Type-specific error handling
                    when (state.type) {
                        HomeViewModel.HomeUiState.ErrorType.NETWORK -> {
                            binding.errorTitle.text = "Connection Error"
                            binding.errorMessage.text = "Please check your internet connection"
                            binding.retryButton.isVisible = true
                        }
                        HomeViewModel.HomeUiState.ErrorType.AUTH -> {
                            binding.errorTitle.text = "Authentication Required"
                            binding.errorMessage.text = "Please log in again"
                            binding.retryButton.isVisible = false
                            navigateToLogin()
                        }
                        HomeViewModel.HomeUiState.ErrorType.SERVER -> {
                            binding.errorTitle.text = "Server Error"
                            binding.errorMessage.text = "Something went wrong. Please try again later."
                            binding.retryButton.isVisible = true
                        }
                        HomeViewModel.HomeUiState.ErrorType.UNKNOWN -> {
                            binding.errorTitle.text = "Unknown Error"
                            binding.errorMessage.text = state.message
                            binding.retryButton.isVisible = true
                        }
                    }
                }

                is HomeViewModel.HomeUiState.NetworkUnavailable -> {
                    binding.progressBar.isVisible = false
                    binding.contentLayout.isVisible = false
                    binding.errorLayout.isVisible = true

                    binding.errorTitle.text = "Offline"
                    binding.errorMessage.text = "You're currently offline"
                    binding.retryButton.isVisible = true

                    // Show cached data if available
                    showCachedContent()
                }
            }
        }
    }
}
```

---

## Android-Specific Error Handling

### Network Error Handling

```kotlin
class NetworkErrorHandler {
    fun handleRetrofitException(exception: Throwable): NetworkError {
        return when (exception) {
            is HttpException -> {
                when (exception.code()) {
                    400 -> NetworkError.BadRequest(exception.message())
                    401 -> NetworkError.Unauthorized
                    403 -> NetworkError.Forbidden
                    404 -> NetworkError.NotFound
                    429 -> NetworkError.TooManyRequests
                    in 500..599 -> NetworkError.ServerError(exception.code())
                    else -> NetworkError.HttpError(exception.code(), exception.message())
                }
            }
            is ConnectException -> NetworkError.NoConnection
            is SocketTimeoutException -> NetworkError.Timeout
            is UnknownHostException -> NetworkError.NoConnection
            is SSLException -> NetworkError.SslError
            is JsonSyntaxException -> NetworkError.ParseError("Invalid response format")
            else -> NetworkError.Unknown(exception.message ?: "Unknown error")
        }
    }
}

sealed class NetworkError : Exception() {
    object NoConnection : NetworkError()
    object Timeout : NetworkError()
    object Unauthorized : NetworkError()
    object Forbidden : NetworkError()
    object NotFound : NetworkError()
    object TooManyRequests : NetworkError()
    data class BadRequest(val details: String?) : NetworkError()
    data class ServerError(val code: Int) : NetworkError()
    data class HttpError(val code: Int, val details: String?) : NetworkError()
    data class ParseError(val details: String) : NetworkError()
    data class SslError(val details: String? = null) : NetworkError()
    data class Unknown(val details: String) : NetworkError()
}

// Usage in Repository
class UserRepository(private val apiService: ApiService) {
    suspend fun loginUser(credentials: LoginCredentials): Result<AuthResponse> {
        return try {
            val response = apiService.login(credentials)
            Result.success(response)
        } catch (e: Exception) {
            val networkError = NetworkErrorHandler().handleRetrofitException(e)
            Result.failure(networkError)
        }
    }
}
```

### Database Error Handling (Room)

```kotlin
@Dao
interface UserDao {
    @Query("SELECT * FROM users WHERE id = :userId")
    suspend fun getUser(userId: String): User?

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertUser(user: User): Long

    @Update
    suspend fun updateUser(user: User): Int

    @Delete
    suspend fun deleteUser(user: User): Int
}

class UserLocalRepository(private val userDao: UserDao) {
    suspend fun saveUser(user: User): Result<Unit> {
        return try {
            val id = userDao.insertUser(user)
            if (id > 0) {
                Result.success(Unit)
            } else {
                Result.failure(DatabaseException("Failed to insert user"))
            }
        } catch (e: SQLiteConstraintException) {
            Result.failure(DatabaseException("User already exists or constraint violation", e))
        } catch (e: SQLiteException) {
            Result.failure(DatabaseException("Database operation failed", e))
        } catch (e: Exception) {
            Result.failure(DatabaseException("Unexpected database error", e))
        }
    }

    suspend fun getUser(userId: String): Result<User> {
        return try {
            val user = userDao.getUser(userId)
            if (user != null) {
                Result.success(user)
            } else {
                Result.failure(UserNotFoundException("User with id $userId not found"))
            }
        } catch (e: Exception) {
            Result.failure(DatabaseException("Failed to retrieve user", e))
        }
    }
}

class DatabaseException(message: String, cause: Throwable? = null) : Exception(message, cause)
class UserNotFoundException(message: String) : Exception(message)
```

### Permission Error Handling

```kotlin
class PermissionHandler(private val activity: ComponentActivity) {
    sealed class PermissionResult {
        object Granted : PermissionResult()
        object Denied : PermissionResult()
        object PermanentlyDenied : PermissionResult()
        data class Error(val exception: Exception) : PermissionResult()
    }

    fun checkCameraPermission(): PermissionResult {
        return when {
            ContextCompat.checkSelfPermission(
                activity,
                Manifest.permission.CAMERA
            ) == PackageManager.PERMISSION_GRANTED -> {
                PermissionResult.Granted
            }
            ActivityCompat.shouldShowRequestPermissionRationale(
                activity,
                Manifest.permission.CAMERA
            ) -> {
                PermissionResult.Denied
            }
            else -> {
                PermissionResult.PermanentlyDenied
            }
        }
    }

    fun handleCameraPermissionResult(result: PermissionResult) {
        when (result) {
            is PermissionResult.Granted -> {
                openCamera()
            }
            is PermissionResult.Denied -> {
                showPermissionRationale()
            }
            is PermissionResult.PermanentlyDenied -> {
                showGoToSettingsDialog()
            }
            is PermissionResult.Error -> {
                showError("Permission error: ${result.exception.message}")
            }
        }
    }

    private fun showGoToSettingsDialog() {
        AlertDialog.Builder(activity)
            .setTitle("Camera Permission Required")
            .setMessage("Camera permission is permanently denied. Please enable it in settings.")
            .setPositiveButton("Go to Settings") { _, _ ->
                openAppSettings()
            }
            .setNegativeButton("Cancel", null)
            .show()
    }

    private fun openAppSettings() {
        try {
            val intent = Intent(Settings.ACTION_APPLICATION_DETAILS_SETTINGS).apply {
                data = Uri.fromParts("package", activity.packageName, null)
            }
            activity.startActivity(intent)
        } catch (e: Exception) {
            showError("Cannot open settings: ${e.message}")
        }
    }
}
```

### Lifecycle Error Handling

```kotlin
abstract class BaseFragment : Fragment() {
    protected val viewLifecycleScope: CoroutineScope
        get() = viewLifecycleOwner.lifecycleScope

    protected fun <T> LiveData<T>.observeSafely(observer: (T) -> Unit) {
        observe(viewLifecycleOwner) { data ->
            try {
                if (isAdded && !isDetached && view != null) {
                    observer(data)
                }
            } catch (e: Exception) {
                handleLifecycleError(e)
            }
        }
    }

    protected fun launchSafely(
        context: CoroutineContext = EmptyCoroutineContext,
        block: suspend CoroutineScope.() -> Unit
    ): Job {
        return viewLifecycleScope.launch(context + CoroutineExceptionHandler { _, throwable ->
            handleCoroutineException(throwable)
        }) {
            try {
                block()
            } catch (e: Exception) {
                handleCoroutineException(e)
            }
        }
    }

    private fun handleLifecycleError(exception: Exception) {
        when (exception) {
            is IllegalStateException -> {
                // Fragment not attached or view destroyed
                Log.w(tag, "Fragment lifecycle error", exception)
            }
            else -> {
                Log.e(tag, "Unexpected lifecycle error", exception)
                // Report to crash reporting service
                FirebaseCrashlytics.getInstance().recordException(exception)
            }
        }
    }

    private fun handleCoroutineException(throwable: Throwable) {
        Log.e(tag, "Coroutine exception", throwable)

        // Handle specific exceptions
        when (throwable) {
            is CancellationException -> {
                // Normal cancellation, don't report
                return
            }
            is NetworkException -> {
                showNetworkError(throwable.message)
            }
            else -> {
                showGenericError("An error occurred: ${throwable.message}")
                FirebaseCrashlytics.getInstance().recordException(throwable)
            }
        }
    }

    protected open fun showNetworkError(message: String?) {
        // Override in subclasses for specific handling
    }

    protected open fun showGenericError(message: String) {
        // Override in subclasses for specific handling
    }

    private val tag: String
        get() = this::class.java.simpleName
}
```

---

## Coroutines Error Handling

### Structured Concurrency and Exception Propagation

```kotlin
class DataSyncService(
    private val userApi: UserApi,
    private val settingsApi: SettingsApi,
    private val localDatabase: AppDatabase
) {
    suspend fun syncUserData(userId: String): Result<SyncResult> = coroutineScope {
        try {
            // Launch multiple concurrent operations
            val userDeferred = async { userApi.getUser(userId) }
            val settingsDeferred = async { settingsApi.getUserSettings(userId) }
            val profileDeferred = async { userApi.getUserProfile(userId) }

            // Await all results - if any fails, all are cancelled
            val user = userDeferred.await()
            val settings = settingsDeferred.await()
            val profile = profileDeferred.await()

            // Save to local database
            localDatabase.withTransaction {
                localDatabase.userDao().insertUser(user)
                localDatabase.settingsDao().insertSettings(settings)
                localDatabase.profileDao().insertProfile(profile)
            }

            Result.success(SyncResult.Success(user, settings, profile))

        } catch (e: CancellationException) {
            // Re-throw cancellation to maintain structured concurrency
            throw e
        } catch (e: Exception) {
            Result.failure(SyncException("Failed to sync user data", e))
        }
    }

    // Use supervisorScope when you want some operations to continue despite failures
    suspend fun syncOptionalData(userId: String): SyncStatus = supervisorScope {
        val results = mutableMapOf<String, Result<Any>>()

        // These operations run independently
        val avatarJob = launch {
            results["avatar"] = try {
                Result.success(userApi.getUserAvatar(userId))
            } catch (e: Exception) {
                Result.failure(e)
            }
        }

        val friendsJob = launch {
            results["friends"] = try {
                Result.success(userApi.getUserFriends(userId))
            } catch (e: Exception) {
                Result.failure(e)
            }
        }

        val achievementsJob = launch {
            results["achievements"] = try {
                Result.success(userApi.getUserAchievements(userId))
            } catch (e: Exception) {
                Result.failure(e)
            }
        }

        // Wait for all to complete
        joinAll(avatarJob, friendsJob, achievementsJob)

        SyncStatus(
            avatarSynced = results["avatar"]?.isSuccess ?: false,
            friendsSynced = results["friends"]?.isSuccess ?: false,
            achievementsSynced = results["achievements"]?.isSuccess ?: false,
            errors = results.values.mapNotNull { it.exceptionOrNull() }
        )
    }
}
```

### CoroutineExceptionHandler

```kotlin
class BackgroundTaskManager(
    private val scope: CoroutineScope
) {
    // Global exception handler for background tasks
    private val exceptionHandler = CoroutineExceptionHandler { _, exception ->
        when (exception) {
            is CancellationException -> {
                // Normal cancellation, don't report
                Log.d(TAG, "Background task was cancelled")
            }
            is NetworkException -> {
                Log.w(TAG, "Network error in background task", exception)
                // Maybe retry later
                scheduleRetry()
            }
            is SecurityException -> {
                Log.e(TAG, "Security error in background task", exception)
                // Handle security issues
                handleSecurityError(exception)
            }
            else -> {
                Log.e(TAG, "Unexpected error in background task", exception)
                // Report to crash reporting
                FirebaseCrashlytics.getInstance().recordException(exception)
            }
        }
    }

    fun startPeriodicSync() {
        scope.launch(exceptionHandler) {
            while (isActive) {
                try {
                    performSync()
                    delay(TimeUnit.HOURS.toMillis(1)) // Sync every hour
                } catch (e: Exception) {
                    // Exception will be handled by the handler above
                    throw e
                }
            }
        }
    }

    private suspend fun performSync() {
        // Sync implementation
    }

    companion object {
        private const val TAG = "BackgroundTaskManager"
    }
}

// Application-level coroutine scope with exception handling
class MyApplication : Application() {
    private val applicationScope = CoroutineScope(
        SupervisorJob() + Dispatchers.Main + CoroutineExceptionHandler { _, throwable ->
            // Global application-level exception handler
            Log.e("MyApplication", "Uncaught coroutine exception", throwable)
            FirebaseCrashlytics.getInstance().recordException(throwable)
        }
    )

    override fun onCreate() {
        super.onCreate()

        // Start background services
        BackgroundTaskManager(applicationScope).startPeriodicSync()
    }
}
```

### Cancellation vs Exceptions

```kotlin
class SearchRepository(
    private val searchApi: SearchApi,
    private val scope: CoroutineScope
) {
    private var currentSearchJob: Job? = null

    suspend fun search(query: String): Result<List<SearchResult>> {
        // Cancel previous search
        currentSearchJob?.cancel()

        currentSearchJob = scope.launch {
            try {
                val results = searchApi.search(query)
                // Process results...
            } catch (e: CancellationException) {
                // Search was cancelled, this is normal
                Log.d(TAG, "Search cancelled for query: $query")
                throw e // Re-throw to maintain cancellation semantics
            } catch (e: Exception) {
                // Actual error occurred
                Log.e(TAG, "Search failed for query: $query", e)
                throw SearchException("Search failed", e)
            }
        }

        return try {
            currentSearchJob?.join()
            // Get results...
            Result.success(emptyList()) // Placeholder
        } catch (e: CancellationException) {
            Result.failure(SearchCancelledException())
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    fun cancelSearch() {
        currentSearchJob?.cancel()
        currentSearchJob = null
    }
}

class SearchCancelledException : CancellationException("Search was cancelled")
class SearchException(message: String, cause: Throwable) : Exception(message, cause)
```

---

## Reactive Error Handling

### Flow Error Handling Operators

```kotlin
class NewsRepository(
    private val newsApi: NewsApi,
    private val newsDao: NewsDao
) {
    // Comprehensive Flow error handling
    fun getNewsFlow(): Flow<List<Article>> = flow {
        emit(newsApi.getLatestNews())
    }.catch { exception ->
        // Handle upstream errors
        when (exception) {
            is NetworkException -> {
                Log.w(TAG, "Network error, falling back to cached data", exception)
                emitAll(newsDao.getCachedNews())
            }
            is HttpException -> {
                Log.e(TAG, "API error: ${exception.code()}", exception)
                throw NewsApiException("Failed to fetch news", exception)
            }
            else -> {
                Log.e(TAG, "Unexpected error", exception)
                throw exception
            }
        }
    }.retry(retries = 3) { exception ->
        // Retry policy
        when (exception) {
            is NetworkException -> {
                delay(1000) // Wait 1 second before retry
                true
            }
            is TimeoutException -> {
                delay(2000) // Wait 2 seconds for timeout
                true
            }
            else -> false // Don't retry for other exceptions
        }
    }.flowOn(Dispatchers.IO)

    // Advanced retry with exponential backoff
    fun getNewsWithExponentialBackoff(): Flow<List<Article>> = flow {
        emit(newsApi.getLatestNews())
    }.retryWhen { exception, attempt ->
        when {
            exception is NetworkException && attempt < 3 -> {
                val delayMs = (1000 * (2.0.pow(attempt.toInt()))).toLong()
                delay(delayMs)
                true
            }
            else -> false
        }
    }.catch { exception ->
        // Final error handling after retries exhausted
        Log.e(TAG, "Failed to fetch news after retries", exception)
        emitAll(newsDao.getCachedNews())
    }
}

// StateFlow error handling
class UserProfileViewModel(
    private val userRepository: UserRepository
) : ViewModel() {
    private val _userState = MutableStateFlow<UserState>(UserState.Loading)
    val userState: StateFlow<UserState> = _userState.asStateFlow()

    fun loadUser(userId: String) {
        viewModelScope.launch {
            userRepository.getUserFlow(userId)
                .catch { exception ->
                    _userState.value = UserState.Error(
                        message = exception.message ?: "Failed to load user",
                        isRetryable = exception is NetworkException
                    )
                }
                .collect { user ->
                    _userState.value = if (user != null) {
                        UserState.Success(user)
                    } else {
                        UserState.Error("User not found", false)
                    }
                }
        }
    }
}

sealed class UserState {
    object Loading : UserState()
    data class Success(val user: User) : UserState()
    data class Error(val message: String, val isRetryable: Boolean) : UserState()
}
```

### SharedFlow and Hot Stream Error Handling

```kotlin
class EventBus {
    private val _events = MutableSharedFlow<AppEvent>()
    val events: SharedFlow<AppEvent> = _events.asSharedFlow()

    // SharedFlow doesn't have catch operator, handle errors at emission
    fun emitEvent(event: AppEvent) {
        try {
            _events.tryEmit(event)
        } catch (e: Exception) {
            Log.e(TAG, "Failed to emit event: $event", e)
            // Could emit an error event instead
            _events.tryEmit(AppEvent.Error("Failed to emit event", e))
        }
    }
}

class NetworkConnectivityManager(
    private val connectivityManager: ConnectivityManager
) {
    private val _networkState = MutableStateFlow<NetworkState>(NetworkState.Unknown)
    val networkState: StateFlow<NetworkState> = _networkState.asStateFlow()

    fun startNetworkMonitoring() {
        try {
            val networkCallback = object : ConnectivityManager.NetworkCallback() {
                override fun onAvailable(network: Network) {
                    _networkState.value = NetworkState.Available
                }

                override fun onLost(network: Network) {
                    _networkState.value = NetworkState.Lost
                }

                override fun onCapabilitiesChanged(
                    network: Network,
                    networkCapabilities: NetworkCapabilities
                ) {
                    val hasInternet = networkCapabilities.hasCapability(
                        NetworkCapabilities.NET_CAPABILITY_INTERNET
                    )
                    _networkState.value = if (hasInternet) {
                        NetworkState.Available
                    } else {
                        NetworkState.Limited
                    }
                }

                override fun onUnavailable() {
                    _networkState.value = NetworkState.Unavailable
                }
            }

            connectivityManager.registerDefaultNetworkCallback(networkCallback)

        } catch (e: SecurityException) {
            Log.e(TAG, "Permission required for network monitoring", e)
            _networkState.value = NetworkState.PermissionDenied
        } catch (e: Exception) {
            Log.e(TAG, "Failed to start network monitoring", e)
            _networkState.value = NetworkState.Error(e)
        }
    }
}

sealed class NetworkState {
    object Unknown : NetworkState()
    object Available : NetworkState()
    object Limited : NetworkState()
    object Lost : NetworkState()
    object Unavailable : NetworkState()
    object PermissionDenied : NetworkState()
    data class Error(val exception: Exception) : NetworkState()
}
```

---

## Functional Error Handling

### Arrow Either Type

```kotlin
// Add Arrow dependency: implementation "io.arrow-kt:arrow-core:1.2.0"
import arrow.core.Either
import arrow.core.left
import arrow.core.right

// Define error types
sealed class ValidationError {
    object EmptyEmail : ValidationError()
    object InvalidEmailFormat : ValidationError()
    object EmptyPassword : ValidationError()
    object WeakPassword : ValidationError()
}

class UserValidator {
    fun validateEmail(email: String): Either<ValidationError, String> {
        return when {
            email.isBlank() -> ValidationError.EmptyEmail.left()
            !Patterns.EMAIL_ADDRESS.matcher(email).matches() -> ValidationError.InvalidEmailFormat.left()
            else -> email.right()
        }
    }

    fun validatePassword(password: String): Either<ValidationError, String> {
        return when {
            password.isBlank() -> ValidationError.EmptyPassword.left()
            password.length < 8 -> ValidationError.WeakPassword.left()
            else -> password.right()
        }
    }

    // Combine validations
    fun validateRegistration(email: String, password: String): Either<List<ValidationError>, Pair<String, String>> {
        val emailValidation = validateEmail(email)
        val passwordValidation = validatePassword(password)

        return when {
            emailValidation.isLeft() && passwordValidation.isLeft() -> {
                listOf(emailValidation.leftOrNull()!!, passwordValidation.leftOrNull()!!).left()
            }
            emailValidation.isLeft() -> listOf(emailValidation.leftOrNull()!!).left()
            passwordValidation.isLeft() -> listOf(passwordValidation.leftOrNull()!!).left()
            else -> Pair(emailValidation.getOrNull()!!, passwordValidation.getOrNull()!!).right()
        }
    }
}

// Usage in ViewModel
class RegistrationViewModel(
    private val userValidator: UserValidator,
    private val authRepository: AuthRepository
) : ViewModel() {

    fun registerUser(email: String, password: String) {
        userValidator.validateRegistration(email, password)
            .fold(
                ifLeft = { errors ->
                    // Handle validation errors
                    _uiState.value = RegistrationUiState.ValidationError(errors)
                },
                ifRight = { (validEmail, validPassword) ->
                    // Proceed with registration
                    performRegistration(validEmail, validPassword)
                }
            )
    }

    private fun performRegistration(email: String, password: String) {
        viewModelScope.launch {
            _uiState.value = RegistrationUiState.Loading

            authRepository.register(email, password)
                .fold(
                    ifLeft = { error ->
                        _uiState.value = RegistrationUiState.RegistrationError(error)
                    },
                    ifRight = { user ->
                        _uiState.value = RegistrationUiState.Success(user)
                    }
                )
        }
    }
}
```

### Railway-Oriented Programming

```kotlin
// Chain operations that can fail
class UserOnboardingService(
    private val userValidator: UserValidator,
    private val authService: AuthService,
    private val profileService: ProfileService,
    private val notificationService: NotificationService
) {
    suspend fun onboardUser(
        email: String,
        password: String,
        profile: UserProfile
    ): Either<OnboardingError, OnboardingResult> {
        return userValidator.validateEmail(email)
            .mapLeft { OnboardingError.ValidationError(it) }
            .flatMap { validEmail ->
                userValidator.validatePassword(password)
                    .mapLeft { OnboardingError.ValidationError(it) }
                    .map { validPassword -> validEmail to validPassword }
            }
            .flatMap { (validEmail, validPassword) ->
                authService.createAccount(validEmail, validPassword)
                    .mapLeft { OnboardingError.AccountCreationError(it) }
            }
            .flatMap { user ->
                profileService.createProfile(user.id, profile)
                    .mapLeft { OnboardingError.ProfileCreationError(it) }
                    .map { user to it }
            }
            .flatMap { (user, profile) ->
                notificationService.sendWelcomeEmail(user.email)
                    .mapLeft { OnboardingError.NotificationError(it) }
                    .map { OnboardingResult(user, profile) }
            }
    }
}

sealed class OnboardingError {
    data class ValidationError(val error: ValidationError) : OnboardingError()
    data class AccountCreationError(val error: AuthError) : OnboardingError()
    data class ProfileCreationError(val error: ProfileError) : OnboardingError()
    data class NotificationError(val error: NotificationError) : OnboardingError()
}

data class OnboardingResult(val user: User, val profile: UserProfile)
```

### Monadic Error Composition

```kotlin
// Extension functions for Result chaining
inline fun <T, R> Result<T>.flatMap(transform: (T) -> Result<R>): Result<R> {
    return when {
        isSuccess -> transform(getOrThrow())
        else -> Result.failure(exceptionOrNull()!!)
    }
}

inline fun <T> Result<T>.recover(transform: (Throwable) -> T): Result<T> {
    return when {
        isSuccess -> this
        else -> Result.success(transform(exceptionOrNull()!!))
    }
}

// Usage example
class DataProcessingService {
    suspend fun processUserData(userId: String): Result<ProcessedData> {
        return fetchUserData(userId)
            .flatMap { userData -> validateData(userData) }
            .flatMap { validData -> enrichData(validData) }
            .flatMap { enrichedData -> transformData(enrichedData) }
            .recover { exception ->
                // Fallback processing
                when (exception) {
                    is NetworkException -> getDefaultProcessedData()
                    else -> throw exception
                }
            }
    }

    private suspend fun fetchUserData(userId: String): Result<UserData> = TODO()
    private suspend fun validateData(data: UserData): Result<ValidatedData> = TODO()
    private suspend fun enrichData(data: ValidatedData): Result<EnrichedData> = TODO()
    private suspend fun transformData(data: EnrichedData): Result<ProcessedData> = TODO()
    private fun getDefaultProcessedData(): ProcessedData = TODO()
}
```

---

## Error Recovery Strategies

### Retry Patterns with Exponential Backoff

```kotlin
class RetryableApiClient(private val apiService: ApiService) {

    // Simple retry with linear backoff
    suspend fun <T> retryWithLinearBackoff(
        maxAttempts: Int = 3,
        delayMs: Long = 1000L,
        operation: suspend () -> T
    ): Result<T> {
        var lastException: Exception? = null

        repeat(maxAttempts) { attempt ->
            try {
                return Result.success(operation())
            } catch (e: Exception) {
                lastException = e

                // Don't retry certain types of errors
                if (!isRetryableException(e) || attempt == maxAttempts - 1) {
                    return Result.failure(e)
                }

                delay(delayMs * (attempt + 1))
            }
        }

        return Result.failure(lastException ?: Exception("Unknown error"))
    }

    // Exponential backoff with jitter
    suspend fun <T> retryWithExponentialBackoff(
        maxAttempts: Int = 5,
        initialDelayMs: Long = 1000L,
        maxDelayMs: Long = 30000L,
        backoffMultiplier: Double = 2.0,
        jitterFactor: Double = 0.1,
        operation: suspend () -> T
    ): Result<T> {
        var lastException: Exception? = null
        var currentDelay = initialDelayMs

        repeat(maxAttempts) { attempt ->
            try {
                return Result.success(operation())
            } catch (e: Exception) {
                lastException = e

                if (!isRetryableException(e) || attempt == maxAttempts - 1) {
                    return Result.failure(e)
                }

                // Add jitter to prevent thundering herd
                val jitter = currentDelay * jitterFactor * Random.nextDouble()
                val delayWithJitter = currentDelay + jitter
                delay(delayWithJitter.toLong())

                // Calculate next delay
                currentDelay = minOf(
                    (currentDelay * backoffMultiplier).toLong(),
                    maxDelayMs
                )
            }
        }

        return Result.failure(lastException ?: Exception("Unknown error"))
    }

    private fun isRetryableException(exception: Exception): Boolean {
        return when (exception) {
            is ConnectException -> true
            is SocketTimeoutException -> true
            is UnknownHostException -> true
            is HttpException -> exception.code() in 500..599 || exception.code() == 429
            else -> false
        }
    }
}

// Usage in Repository
class UserRepository(private val retryableClient: RetryableApiClient) {
    suspend fun fetchCriticalUserData(userId: String): Result<User> {
        return retryableClient.retryWithExponentialBackoff(
            maxAttempts = 5,
            initialDelayMs = 1000L
        ) {
            apiService.getCriticalUserData(userId)
        }
    }
}
```

### Circuit Breaker Pattern

```kotlin
class CircuitBreaker(
    private val failureThreshold: Int = 5,
    private val recoveryTimeout: Long = 30000L, // 30 seconds
    private val successThreshold: Int = 3
) {
    private var state: CircuitState = CircuitState.Closed
    private var failureCount = 0
    private var lastFailureTime = 0L
    private var successCount = 0

    sealed class CircuitState {
        object Closed : CircuitState()    // Normal operation
        object Open : CircuitState()      // Failing fast
        object HalfOpen : CircuitState()  // Testing recovery
    }

    class CircuitBreakerOpenException : Exception("Circuit breaker is open")

    suspend fun <T> execute(operation: suspend () -> T): T {
        when (state) {
            CircuitState.Open -> {
                if (System.currentTimeMillis() - lastFailureTime >= recoveryTimeout) {
                    state = CircuitState.HalfOpen
                    successCount = 0
                } else {
                    throw CircuitBreakerOpenException()
                }
            }
            else -> { /* Continue */ }
        }

        return try {
            val result = operation()
            onSuccess()
            result
        } catch (e: Exception) {
            onFailure()
            throw e
        }
    }

    private fun onSuccess() {
        when (state) {
            CircuitState.Closed -> {
                failureCount = 0
            }
            CircuitState.HalfOpen -> {
                successCount++
                if (successCount >= successThreshold) {
                    state = CircuitState.Closed
                    failureCount = 0
                }
            }
            else -> { /* No action */ }
        }
    }

    private fun onFailure() {
        failureCount++
        lastFailureTime = System.currentTimeMillis()

        if (failureCount >= failureThreshold) {
            state = CircuitState.Open
        }
    }
}

// Usage in service
class ExternalApiService(private val circuitBreaker: CircuitBreaker) {
    suspend fun callExternalApi(): Result<ApiResponse> {
        return try {
            val response = circuitBreaker.execute {
                // Actual API call
                httpClient.get("https://external-api.com/data")
            }
            Result.success(response)
        } catch (e: CircuitBreaker.CircuitBreakerOpenException) {
            // Circuit is open, return cached data or error
            Result.failure(ServiceUnavailableException("External service is temporarily unavailable"))
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}
```

### Fallback Mechanisms

```kotlin
class DataService(
    private val primaryDataSource: PrimaryDataSource,
    private val secondaryDataSource: SecondaryDataSource,
    private val cacheDataSource: CacheDataSource,
    private val defaultDataSource: DefaultDataSource
) {
    suspend fun getData(id: String): Result<Data> {
        // Try primary source
        primaryDataSource.getData(id)
            .onSuccess { data ->
                // Cache successful result
                cacheDataSource.cacheData(id, data)
                return Result.success(data)
            }

        // Fallback to secondary source
        secondaryDataSource.getData(id)
            .onSuccess { data ->
                cacheDataSource.cacheData(id, data)
                return Result.success(data)
            }

        // Fallback to cache
        cacheDataSource.getCachedData(id)
            .onSuccess { cachedData ->
                return Result.success(cachedData)
            }

        // Final fallback to default
        return try {
            val defaultData = defaultDataSource.getDefaultData(id)
            Result.success(defaultData)
        } catch (e: Exception) {
            Result.failure(DataUnavailableException("All data sources failed", e))
        }
    }
}

// Offline-first pattern
class OfflineFirstRepository(
    private val localDao: DataDao,
    private val remoteApi: DataApi,
    private val networkManager: NetworkManager
) {
    fun getDataStream(id: String): Flow<Resource<Data>> = flow {
        // Always emit cached data first
        val cachedData = localDao.getData(id)
        if (cachedData != null) {
            emit(Resource.Success(cachedData, fromCache = true))
        }

        // Try to fetch fresh data if online
        if (networkManager.isOnline()) {
            try {
                val freshData = remoteApi.getData(id)
                localDao.insertData(freshData)
                emit(Resource.Success(freshData, fromCache = false))
            } catch (e: Exception) {
                if (cachedData == null) {
                    emit(Resource.Error(e))
                }
                // If we have cached data, we already emitted it, so just log the error
                Log.w(TAG, "Failed to fetch fresh data, using cached version", e)
            }
        } else {
            if (cachedData == null) {
                emit(Resource.Error(NetworkUnavailableException()))
            }
        }
    }
}

data class Resource<T>(
    val data: T?,
    val error: Throwable?,
    val fromCache: Boolean = false
) {
    companion object {
        fun <T> Success(data: T, fromCache: Boolean = false) = Resource(data, null, fromCache)
        fun <T> Error(error: Throwable) = Resource<T>(null, error)
    }

    val isSuccess: Boolean get() = error == null && data != null
    val isError: Boolean get() = error != null
}
```

---

## Testing Error Handling

### Unit Testing Exceptions

```kotlin
class UserRepositoryTest {
    @Mock private lateinit var apiService: ApiService
    @Mock private lateinit var userDao: UserDao

    private lateinit var userRepository: UserRepository

    @Before
    fun setup() {
        MockitoAnnotations.openMocks(this)
        userRepository = UserRepository(apiService, userDao)
    }

    @Test
    fun `getUser should throw NetworkException when API call fails`() = runTest {
        // Given
        val userId = "123"
        val networkError = ConnectException("Network error")
        whenever(apiService.getUser(userId)).thenThrow(networkError)

        // When & Then
        val exception = assertThrows<NetworkException> {
            userRepository.getUser(userId)
        }

        assertEquals("Network error", exception.message)
        assertEquals(networkError, exception.cause)
    }

    @Test
    fun `getUser should return Result failure when API call fails`() = runTest {
        // Given
        val userId = "123"
        whenever(apiService.getUser(userId)).thenThrow(ConnectException("No connection"))

        // When
        val result = userRepository.getUserResult(userId)

        // Then
        assertTrue(result.isFailure)
        assertTrue(result.exceptionOrNull() is NetworkException)
    }

    @Test
    fun `getUser should fallback to cache when network fails`() = runTest {
        // Given
        val userId = "123"
        val cachedUser = User(userId, "Cached User")
        whenever(apiService.getUser(userId)).thenThrow(ConnectException())
        whenever(userDao.getUser(userId)).thenReturn(cachedUser)

        // When
        val result = userRepository.getUserWithFallback(userId)

        // Then
        assertTrue(result.isSuccess)
        assertEquals(cachedUser, result.getOrNull())
        verify(userDao).getUser(userId)
    }
}

// Testing sealed class error states
class UserViewModelTest {
    @Mock private lateinit var userRepository: UserRepository
    private lateinit var viewModel: UserViewModel

    @Before
    fun setup() {
        MockitoAnnotations.openMocks(this)
        viewModel = UserViewModel(userRepository)
    }

    @Test
    fun `loadUser should emit Error state when repository throws exception`() = runTest {
        // Given
        val userId = "123"
        val exception = NetworkException("Network error")
        whenever(userRepository.getUser(userId)).thenThrow(exception)

        // When
        viewModel.loadUser(userId)

        // Then
        val uiState = viewModel.uiState.value
        assertTrue(uiState is UserUiState.Error)
        assertEquals("Network error", (uiState as UserUiState.Error).message)
    }

    @Test
    fun `loadUser should handle different error types correctly`() = runTest {
        val testCases = listOf(
            NetworkException("Network") to UserUiState.ErrorType.NETWORK,
            SecurityException("Auth") to UserUiState.ErrorType.SECURITY,
            IllegalArgumentException("Validation") to UserUiState.ErrorType.VALIDATION
        )

        testCases.forEach { (exception, expectedType) ->
            // Given
            whenever(userRepository.getUser(any())).thenThrow(exception)

            // When
            viewModel.loadUser("123")

            // Then
            val uiState = viewModel.uiState.value as UserUiState.Error
            assertEquals(expectedType, uiState.type)
        }
    }
}
```

### Integration Testing Error Scenarios

```kotlin
@RunWith(AndroidJUnit4::class)
@SmallTest
class UserDatabaseTest {
    @get:Rule
    val instantExecutorRule = InstantTaskExecutorRule()

    private lateinit var database: TestDatabase
    private lateinit var userDao: UserDao

    @Before
    fun setup() {
        database = Room.inMemoryDatabaseBuilder(
            ApplicationProvider.getApplicationContext(),
            TestDatabase::class.java
        ).allowMainThreadQueries().build()

        userDao = database.userDao()
    }

    @After
    fun tearDown() {
        database.close()
    }

    @Test
    fun insertUser_constraintViolation_throwsException() = runTest {
        // Given
        val user1 = User("123", "user@example.com", "John Doe")
        val user2 = User("123", "different@example.com", "Jane Doe") // Same ID

        // When
        userDao.insertUser(user1)

        // Then
        assertThrows<SQLiteConstraintException> {
            runBlocking { userDao.insertUser(user2) }
        }
    }

    @Test
    fun getUserById_nonExistentId_returnsNull() = runTest {
        // When
        val result = userDao.getUserById("non-existent")

        // Then
        assertNull(result)
    }
}

// Network error simulation in tests
@RunWith(RobolectricTestRunner::class)
class NetworkErrorIntegrationTest {

    @Test
    fun `api call with network error should fallback correctly`() = runTest {
        // Given
        val mockServer = MockWebServer()
        mockServer.enqueue(MockResponse().setSocketPolicy(SocketPolicy.DISCONNECT_DURING_REQUEST_BODY))

        val retrofit = Retrofit.Builder()
            .baseUrl(mockServer.url("/"))
            .addConverterFactory(GsonConverterFactory.create())
            .build()

        val apiService = retrofit.create(ApiService::class.java)
        val repository = UserRepository(apiService, mockUserDao)

        // When & Then
        val exception = assertThrows<NetworkException> {
            runBlocking { repository.getUser("123") }
        }

        assertTrue(exception.cause is ConnectException)
        mockServer.shutdown()
    }
}
```

### UI Testing Error States

```kotlin
@RunWith(AndroidJUnit4::class)
@LargeTest
class UserProfileActivityTest {

    @get:Rule
    val activityRule = ActivityScenarioRule(UserProfileActivity::class.java)

    @Test
    fun displayErrorState_whenUserLoadFails() {
        // Given
        val intent = Intent().apply {
            putExtra("user_id", "123")
            putExtra("simulate_error", true) // Test-only parameter
        }

        // When
        ActivityScenario.launch<UserProfileActivity>(intent)

        // Then
        onView(withId(R.id.error_layout))
            .check(matches(isDisplayed()))

        onView(withId(R.id.error_message))
            .check(matches(withText("Failed to load user profile")))

        onView(withId(R.id.retry_button))
            .check(matches(isDisplayed()))
    }

    @Test
    fun retryButton_clickAfterError_retriesOperation() {
        // Given - Error state is displayed
        val intent = Intent().apply {
            putExtra("user_id", "123")
            putExtra("simulate_error", true)
        }
        ActivityScenario.launch<UserProfileActivity>(intent)

        // When
        onView(withId(R.id.retry_button)).perform(click())

        // Then
        onView(withId(R.id.loading_indicator))
            .check(matches(isDisplayed()))
    }

    @Test
    fun networkErrorDialog_showsCorrectMessage() {
        // Given
        val intent = Intent().apply {
            putExtra("user_id", "123")
            putExtra("error_type", "network")
        }

        ActivityScenario.launch<UserProfileActivity>(intent)

        // Then
        onView(withText("Network Error"))
            .inRoot(isDialog())
            .check(matches(isDisplayed()))

        onView(withText("Please check your internet connection"))
            .inRoot(isDialog())
            .check(matches(isDisplayed()))
    }
}
```

---

## Error Logging and Monitoring

### Firebase Crashlytics Integration

```kotlin
class CrashReportingManager private constructor() {
    companion object {
        @Volatile
        private var INSTANCE: CrashReportingManager? = null

        fun getInstance(): CrashReportingManager {
            return INSTANCE ?: synchronized(this) {
                INSTANCE ?: CrashReportingManager().also { INSTANCE = it }
            }
        }
    }

    private val crashlytics = FirebaseCrashlytics.getInstance()

    fun initialize(userId: String?, userEmail: String?) {
        crashlytics.setUserId(userId ?: "anonymous")
        crashlytics.setCustomKey("user_email", userEmail ?: "unknown")
        crashlytics.setCustomKey("app_version", BuildConfig.VERSION_NAME)
        crashlytics.setCustomKey("build_type", BuildConfig.BUILD_TYPE)
    }

    fun logException(exception: Throwable, context: String? = null) {
        context?.let { crashlytics.setCustomKey("error_context", it) }
        crashlytics.recordException(exception)
    }

    fun logError(message: String, context: Map<String, String> = emptyMap()) {
        context.forEach { (key, value) ->
            crashlytics.setCustomKey(key, value)
        }
        crashlytics.log("ERROR: $message")
    }

    fun logWarning(message: String, context: Map<String, String> = emptyMap()) {
        context.forEach { (key, value) ->
            crashlytics.setCustomKey(key, value)
        }
        crashlytics.log("WARNING: $message")
    }

    fun setBreadcrumb(message: String) {
        crashlytics.log(message)
    }

    fun setCustomAttribute(key: String, value: String) {
        crashlytics.setCustomKey(key, value)
    }
}

// Usage in repositories and services
class UserRepository(
    private val apiService: ApiService,
    private val crashReporting: CrashReportingManager = CrashReportingManager.getInstance()
) {
    suspend fun updateUserProfile(user: User): Result<User> {
        crashReporting.setBreadcrumb("Starting user profile update for ${user.id}")

        return try {
            val updatedUser = apiService.updateUser(user)
            crashReporting.setBreadcrumb("User profile updated successfully")
            Result.success(updatedUser)
        } catch (e: Exception) {
            crashReporting.logException(
                exception = e,
                context = "updateUserProfile"
            )
            crashReporting.setCustomAttribute("user_id", user.id)
            crashReporting.setCustomAttribute("operation", "update_profile")

            Result.failure(e)
        }
    }
}
```

### Structured Logging

```kotlin
object Logger {
    private const val TAG = "MyApp"
    private val crashReporting = CrashReportingManager.getInstance()

    fun d(tag: String, message: String, context: Map<String, Any> = emptyMap()) {
        if (BuildConfig.DEBUG) {
            Log.d(formatTag(tag), formatMessage(message, context))
        }
    }

    fun i(tag: String, message: String, context: Map<String, Any> = emptyMap()) {
        Log.i(formatTag(tag), formatMessage(message, context))
        crashReporting.setBreadcrumb("INFO: $message")
    }

    fun w(tag: String, message: String, context: Map<String, Any> = emptyMap(), throwable: Throwable? = null) {
        Log.w(formatTag(tag), formatMessage(message, context), throwable)
        crashReporting.logWarning(message, context.mapValues { it.value.toString() })
        throwable?.let { crashReporting.logException(it, "warning") }
    }

    fun e(tag: String, message: String, context: Map<String, Any> = emptyMap(), throwable: Throwable? = null) {
        Log.e(formatTag(tag), formatMessage(message, context), throwable)
        crashReporting.logError(message, context.mapValues { it.value.toString() })
        throwable?.let { crashReporting.logException(it, "error") }
    }

    private fun formatTag(tag: String): String = "[$TAG] $tag"

    private fun formatMessage(message: String, context: Map<String, Any>): String {
        return if (context.isEmpty()) {
            message
        } else {
            val contextString = context.entries.joinToString(", ") { "${it.key}=${it.value}" }
            "$message | Context: [$contextString]"
        }
    }
}

// Usage examples
class NetworkManager {
    companion object {
        private const val TAG = "NetworkManager"
    }

    suspend fun makeApiCall(endpoint: String, userId: String): Result<String> {
        Logger.d(TAG, "Starting API call", mapOf(
            "endpoint" to endpoint,
            "user_id" to userId,
            "timestamp" to System.currentTimeMillis()
        ))

        return try {
            val response = httpClient.get(endpoint)
            Logger.i(TAG, "API call successful", mapOf(
                "endpoint" to endpoint,
                "response_code" to 200,
                "response_time" to "150ms"
            ))
            Result.success(response)
        } catch (e: Exception) {
            Logger.e(TAG, "API call failed", mapOf(
                "endpoint" to endpoint,
                "user_id" to userId,
                "error_type" to e::class.simpleName
            ), e)
            Result.failure(e)
        }
    }
}
```

### Performance and Error Monitoring

```kotlin
class PerformanceMonitor {
    private val performanceTracing = FirebasePerformance.getInstance()
    private val crashReporting = CrashReportingManager.getInstance()

    fun <T> measurePerformance(
        operationName: String,
        attributes: Map<String, String> = emptyMap(),
        operation: () -> T
    ): T {
        val trace = performanceTracing.newTrace(operationName)

        // Add custom attributes
        attributes.forEach { (key, value) ->
            trace.putAttribute(key, value)
        }

        trace.start()

        return try {
            val result = operation()
            trace.putAttribute("status", "success")
            result
        } catch (e: Exception) {
            trace.putAttribute("status", "error")
            trace.putAttribute("error_type", e::class.simpleName ?: "Unknown")

            crashReporting.logException(e, operationName)
            throw e
        } finally {
            trace.stop()
        }
    }

    suspend fun <T> measureSuspendingPerformance(
        operationName: String,
        attributes: Map<String, String> = emptyMap(),
        operation: suspend () -> T
    ): T {
        val startTime = System.currentTimeMillis()
        val trace = performanceTracing.newTrace(operationName)

        attributes.forEach { (key, value) ->
            trace.putAttribute(key, value)
        }

        trace.start()

        return try {
            val result = operation()
            trace.putAttribute("status", "success")
            result
        } catch (e: Exception) {
            trace.putAttribute("status", "error")
            trace.putAttribute("error_type", e::class.simpleName ?: "Unknown")
            crashReporting.logException(e, operationName)
            throw e
        } finally {
            val duration = System.currentTimeMillis() - startTime
            trace.putMetric("duration_ms", duration)
            trace.stop()
        }
    }
}

// Usage in repository
class DataRepository(
    private val apiService: ApiService,
    private val performanceMonitor: PerformanceMonitor
) {
    suspend fun fetchUserData(userId: String): Result<UserData> {
        return try {
            val userData = performanceMonitor.measureSuspendingPerformance(
                operationName = "fetch_user_data",
                attributes = mapOf(
                    "user_id" to userId,
                    "data_source" to "api"
                )
            ) {
                apiService.getUserData(userId)
            }
            Result.success(userData)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}
```

---

## Best Practices Summary

### 1. Choose the Right Error Handling Pattern

```kotlin
// Use Result for operations that can fail
suspend fun fetchData(): Result<Data> = TODO()

// Use sealed classes for UI states
sealed class UiState {
    object Loading : UiState()
    data class Success(val data: List<Item>) : UiState()
    data class Error(val message: String) : UiState()
}

// Use exceptions for programming errors
fun validateInput(input: String) {
    require(input.isNotBlank()) { "Input cannot be blank" }
}
```

### 2. Fail Fast and Fail Clearly

```kotlin
class UserService {
    fun processUser(user: User?) {
        // Fail fast with clear error message
        requireNotNull(user) { "User cannot be null" }
        require(user.email.isNotBlank()) { "User email cannot be blank" }
        require(user.id.isNotBlank()) { "User ID cannot be blank" }

        // Process user...
    }
}
```

### 3. Provide Meaningful Error Messages

```kotlin
sealed class AuthError {
    object NetworkUnavailable : AuthError() {
        override fun toString() = "Please check your internet connection and try again"
    }

    data class InvalidCredentials(val field: String) : AuthError() {
        override fun toString() = "Invalid $field. Please check and try again"
    }

    data class AccountLocked(val unlockTime: Long) : AuthError() {
        override fun toString() = "Account locked. Try again after ${formatTime(unlockTime)}"
    }
}
```

### 4. Handle Errors at the Right Level

```kotlin
// Handle technical errors in repositories
class UserRepository {
    suspend fun getUser(id: String): Result<User> {
        return try {
            // Technical error handling
            val user = apiService.getUser(id)
            Result.success(user)
        } catch (e: NetworkException) {
            Result.failure(e)
        }
    }
}

// Handle business logic errors in use cases
class GetUserUseCase {
    suspend operator fun invoke(id: String): UserResult {
        return when (val result = repository.getUser(id)) {
            is Result.Success -> {
                // Business logic validation
                val user = result.getOrThrow()
                if (user.isActive) {
                    UserResult.Success(user)
                } else {
                    UserResult.UserInactive
                }
            }
            is Result.Failure -> UserResult.Error(result.exceptionOrNull()!!)
        }
    }
}

// Handle UI errors in ViewModels/UI layer
class UserViewModel {
    fun loadUser(id: String) {
        viewModelScope.launch {
            when (val result = getUserUseCase(id)) {
                is UserResult.Success -> updateUiForSuccess(result.user)
                is UserResult.UserInactive -> updateUiForInactiveUser()
                is UserResult.Error -> updateUiForError(result.exception)
            }
        }
    }
}
```

This comprehensive guide covers all aspects of error handling in Kotlin mobile development, from basic exception handling to advanced functional patterns and monitoring strategies. The key is to choose the right approach for your specific use case and maintain consistency throughout your application.