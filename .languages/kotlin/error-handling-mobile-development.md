# Comprehensive Error Handling in Kotlin for Mobile Development

> **Last Updated**: September 2024
> **Focus**: Mobile-first error handling patterns for Android applications

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

### 1. Exception Handling

Kotlin's exception handling is built on familiar try-catch-finally patterns but with important differences from Java.

#### Basic Syntax

```kotlin
try {
    // Potentially risky code
    val result = riskyOperation()
    processResult(result)
} catch (e: NetworkException) {
    // Handle specific network errors
    handleNetworkError(e)
} catch (e: Exception) {
    // Handle any other exceptions
    handleGenericError(e)
} finally {
    // Cleanup code that always runs
    cleanup()
}
```

#### Key Characteristics

- **All exceptions are unchecked**: Unlike Java, Kotlin doesn't distinguish between checked and unchecked exceptions
- **Exceptions are objects**: Created from subclasses of the `Exception` class
- **Clean propagation**: Exceptions bubble up naturally without forced handling

#### Mobile Context Example

```kotlin
class UserRepository {
    suspend fun fetchUserData(userId: String): User {
        try {
            val response = apiService.getUser(userId)
            return response.toUser()
        } catch (e: HttpException) {
            when (e.code()) {
                404 -> throw UserNotFoundException("User $userId not found")
                401 -> throw AuthenticationException("Authentication required")
                else -> throw NetworkException("Failed to fetch user data", e)
            }
        } catch (e: IOException) {
            throw ConnectivityException("Network connection failed", e)
        }
    }
}
```

### 2. Custom Exceptions

Creating domain-specific exceptions improves error handling clarity in mobile applications.

```kotlin
// Base exception for the app
sealed class AppException(message: String, cause: Throwable? = null) : Exception(message, cause)

// Network-related exceptions
class NetworkException(message: String, cause: Throwable? = null) : AppException(message, cause)
class ConnectivityException(message: String, cause: Throwable? = null) : NetworkException(message, cause)
class ApiException(val code: Int, message: String) : NetworkException(message)

// Data-related exceptions
class DataException(message: String, cause: Throwable? = null) : AppException(message, cause)
class DatabaseException(message: String, cause: Throwable? = null) : DataException(message, cause)
class CacheException(message: String, cause: Throwable? = null) : DataException(message, cause)

// Business logic exceptions
class ValidationException(val field: String, message: String) : AppException(message)
class AuthenticationException(message: String) : AppException(message)
class AuthorizationException(message: String) : AppException(message)
```

### 3. Precondition Functions

Kotlin provides built-in functions for validating inputs and state:

```kotlin
class UserValidator {
    fun validateUser(user: User) {
        // Throws IllegalArgumentException if condition is false
        require(user.email.isNotEmpty()) { "Email cannot be empty" }
        require(user.age >= 0) { "Age must be non-negative" }

        // Throws IllegalStateException if condition is false
        check(isInitialized) { "Validator must be initialized" }

        // Throws IllegalStateException with custom message
        if (user.isBlocked) {
            error("User is blocked and cannot be processed")
        }
    }
}
```

**Best Practices:**
- Order catch blocks from most specific to least specific
- Use precondition functions for input validation
- Create new exception instances each time (avoid object declarations)
- Include original cause when re-throwing exceptions

---

## Kotlin-Specific Error Handling

### 1. Null Safety

Kotlin's null safety system prevents `NullPointerException` at compile time, making it a cornerstone of error prevention.

#### Nullable vs Non-Nullable Types

```kotlin
class UserProfile {
    // Non-nullable - cannot be null
    val userId: String = "user123"

    // Nullable - can be null
    var displayName: String? = null
    var profilePicture: String? = null

    fun updateProfile(name: String?, picture: String?) {
        // Safe assignment
        displayName = name
        profilePicture = picture
    }
}
```

#### Safe Call Operator

```kotlin
class ProfileManager {
    fun getProfileImageSize(user: User?): Int {
        // Safe call - returns null if user or profilePicture is null
        val size = user?.profilePicture?.length

        // Elvis operator provides default value
        return size ?: 0
    }

    fun displayUserInfo(user: User?) {
        // Chain safe calls
        user?.profile?.displayName?.let { name ->
            // This block only executes if all values are non-null
            showUserName(name)
        }
    }
}
```

#### Safe Casting

```kotlin
fun handleUserInput(input: Any) {
    // Safe cast - returns null if cast fails
    val userString = input as? String

    if (userString != null) {
        processUserString(userString)
    }

    // Alternative with let
    (input as? User)?.let { user ->
        processUser(user)
    }
}
```

#### Mobile-Specific Example

```kotlin
class LocationService {
    private var lastLocation: Location? = null

    fun getCurrentLocationInfo(): String {
        return lastLocation?.let { location ->
            "Lat: ${location.latitude}, Lng: ${location.longitude}"
        } ?: "Location not available"
    }

    fun calculateDistance(destination: Location?): Double? {
        // Safe chaining prevents crashes
        return lastLocation?.let { current ->
            destination?.let { dest ->
                calculateDistanceBetween(current, dest)
            }
        }
    }
}
```

### 2. Result Type

Kotlin's `Result` type provides functional error handling without exceptions.

#### Basic Usage

```kotlin
import kotlin.Result.Companion.failure
import kotlin.Result.Companion.success

class ApiService {
    fun fetchUserSafely(userId: String): Result<User> {
        return runCatching {
            // This could throw exceptions
            val response = httpClient.get("/users/$userId")
            parseUserResponse(response)
        }
    }

    fun processUserResult(result: Result<User>) {
        result
            .onSuccess { user ->
                // Handle successful case
                updateUI(user)
            }
            .onFailure { exception ->
                // Handle failure case
                showError(exception.message)
            }
    }
}
```

#### Chaining Operations

```kotlin
class UserRepository {
    suspend fun getUserWithDetails(userId: String): Result<UserDetails> {
        return runCatching {
            fetchUser(userId)
        }.mapCatching { user ->
            // Transform success value
            enhanceUserDetails(user)
        }.recoverCatching { exception ->
            // Provide fallback for specific errors
            when (exception) {
                is NetworkException -> getCachedUser(userId)
                else -> throw exception
            }
        }
    }

    fun handleUserDetails(userId: String) {
        viewModelScope.launch {
            getUserWithDetails(userId)
                .fold(
                    onSuccess = { userDetails ->
                        _userState.value = UserState.Success(userDetails)
                    },
                    onFailure = { exception ->
                        _userState.value = UserState.Error(exception.message ?: "Unknown error")
                    }
                )
        }
    }
}
```

#### Recovery Patterns

```kotlin
class DataManager {
    fun getDataWithFallback(id: String): Result<Data> {
        return runCatching {
            fetchFromNetwork(id)
        }.recoverCatching { networkError ->
            // Try cache if network fails
            fetchFromCache(id)
        }.recover { cacheError ->
            // Provide default if both fail
            getDefaultData(id)
        }
    }

    fun getDataWithRetry(id: String, maxRetries: Int = 3): Result<Data> {
        var lastException: Throwable? = null

        repeat(maxRetries) { attempt ->
            runCatching {
                fetchFromNetwork(id)
            }.onSuccess { data ->
                return success(data)
            }.onFailure { exception ->
                lastException = exception
                if (attempt < maxRetries - 1) {
                    delay(1000 * (attempt + 1)) // Exponential backoff
                }
            }
        }

        return failure(lastException ?: Exception("Max retries exceeded"))
    }
}
```

**Important Limitations:**
- `runCatching` catches `CancellationException`, which can break structured concurrency in coroutines
- Use with caution in coroutine contexts; prefer proper exception handling

### 3. Elvis Operator

The Elvis operator (`?:`) provides elegant fallback values:

```kotlin
class UserPreferences {
    private val preferences: SharedPreferences

    fun getUserTheme(): String {
        // Return saved theme or default
        return preferences.getString("theme", null) ?: "light"
    }

    fun getUserTimeout(): Long {
        val savedTimeout = preferences.getLong("timeout", -1)
        // Use saved value or calculate default
        return if (savedTimeout > 0) savedTimeout else calculateDefaultTimeout()
    }

    // Chain multiple fallbacks
    fun getUserDisplayName(user: User?): String {
        return user?.displayName
            ?: user?.email?.substringBefore('@')
            ?: user?.username
            ?: "Anonymous User"
    }
}
```

---

## Sealed Classes for Error States

Sealed classes provide type-safe error handling and are perfect for managing UI states in mobile applications.

### 1. Basic UI State Pattern

```kotlin
sealed class UiState<out T> {
    object Loading : UiState<Nothing>()
    data class Success<T>(val data: T) : UiState<T>()
    data class Error(val message: String, val throwable: Throwable? = null) : UiState<Nothing>()
}

// Usage in ViewModel
class UserViewModel : ViewModel() {
    private val _userState = MutableStateFlow<UiState<User>>(UiState.Loading)
    val userState: StateFlow<UiState<User>> = _userState.asStateFlow()

    fun loadUser(userId: String) {
        viewModelScope.launch {
            _userState.value = UiState.Loading

            try {
                val user = userRepository.getUser(userId)
                _userState.value = UiState.Success(user)
            } catch (e: Exception) {
                _userState.value = UiState.Error("Failed to load user", e)
            }
        }
    }
}
```

### 2. Advanced Error States

```kotlin
sealed class NetworkResult<out T> {
    data class Success<T>(val data: T) : NetworkResult<T>()
    data class Loading(val progress: Float = 0f) : NetworkResult<Nothing>()

    sealed class Error : NetworkResult<Nothing>() {
        data class NetworkError(val code: Int, val message: String) : Error()
        data class TimeoutError(val timeoutMs: Long) : Error()
        object NoInternetError : Error()
        data class ValidationError(val errors: Map<String, String>) : Error()
        data class UnknownError(val throwable: Throwable) : Error()
    }
}

// Enhanced error handling
class ApiRepository {
    suspend fun fetchUsers(): NetworkResult<List<User>> {
        return try {
            if (!networkMonitor.isConnected()) {
                return NetworkResult.Error.NoInternetError
            }

            val response = apiService.getUsers()
            when (response.code()) {
                200 -> NetworkResult.Success(response.body()!!)
                400 -> NetworkResult.Error.ValidationError(parseValidationErrors(response))
                408 -> NetworkResult.Error.TimeoutError(30000)
                else -> NetworkResult.Error.NetworkError(response.code(), response.message())
            }
        } catch (e: IOException) {
            NetworkResult.Error.NoInternetError
        } catch (e: HttpException) {
            NetworkResult.Error.NetworkError(e.code(), e.message())
        } catch (e: Exception) {
            NetworkResult.Error.UnknownError(e)
        }
    }
}
```

### 3. Handling States with When Expressions

```kotlin
// UI Layer handling
@Composable
fun UserScreen(viewModel: UserViewModel) {
    val userState by viewModel.userState.collectAsState()

    when (userState) {
        is UiState.Loading -> {
            LoadingIndicator()
        }

        is UiState.Success -> {
            UserContent(user = userState.data)
        }

        is UiState.Error -> {
            ErrorMessage(
                message = userState.message,
                onRetry = { viewModel.retry() }
            )
        }
    }
}

// Repository layer handling
class UserRepository {
    suspend fun processNetworkResult(result: NetworkResult<User>): User {
        return when (result) {
            is NetworkResult.Success -> result.data

            is NetworkResult.Error.NetworkError -> {
                logger.logError("Network error: ${result.code} - ${result.message}")
                throw ApiException(result.code, result.message)
            }

            is NetworkResult.Error.NoInternetError -> {
                logger.logError("No internet connection")
                throw ConnectivityException("No internet connection available")
            }

            is NetworkResult.Error.TimeoutError -> {
                logger.logError("Request timeout: ${result.timeoutMs}ms")
                throw NetworkTimeoutException("Request timed out after ${result.timeoutMs}ms")
            }

            is NetworkResult.Error.ValidationError -> {
                logger.logError("Validation errors: ${result.errors}")
                throw ValidationException(result.errors)
            }

            is NetworkResult.Error.UnknownError -> {
                logger.logError("Unknown error", result.throwable)
                throw result.throwable
            }

            is NetworkResult.Loading -> {
                error("Cannot process loading state")
            }
        }
    }
}
```

### 4. Generic Error Wrapper

```kotlin
sealed class Resource<T>(
    val data: T? = null,
    val message: String? = null,
    val throwable: Throwable? = null
) {
    class Success<T>(data: T) : Resource<T>(data)
    class Error<T>(message: String?, throwable: Throwable? = null, data: T? = null) : Resource<T>(data, message, throwable)
    class Loading<T>(data: T? = null) : Resource<T>(data)
}

// Extension functions for easier handling
fun <T> Resource<T>.onSuccess(action: (T) -> Unit): Resource<T> {
    if (this is Resource.Success) action(data)
    return this
}

fun <T> Resource<T>.onError(action: (String?, Throwable?) -> Unit): Resource<T> {
    if (this is Resource.Error) action(message, throwable)
    return this
}

fun <T> Resource<T>.onLoading(action: () -> Unit): Resource<T> {
    if (this is Resource.Loading) action()
    return this
}

// Usage
class NewsRepository {
    suspend fun getLatestNews(): Resource<List<Article>> {
        return try {
            Resource.Loading()
            val articles = newsApi.getLatestArticles()
            Resource.Success(articles)
        } catch (e: Exception) {
            Resource.Error("Failed to load news", e)
        }
    }
}
```

**Best Practices:**
- Use sealed classes for fixed, finite states
- Make error states as specific as possible
- Include both user-friendly messages and technical details
- Use exhaustive `when` expressions for compile-time safety
- Consider including recovery actions in error states

---

## Android-Specific Error Handling

Mobile applications face unique error scenarios that require specialized handling patterns.

### 1. Network Error Handling

Android restricts network operations on the main thread to prevent UI blocking.

#### Basic Network Error Pattern

```kotlin
class NetworkManager {
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())

    suspend fun makeNetworkCall(): Result<ApiResponse> = withContext(Dispatchers.IO) {
        try {
            if (!isNetworkAvailable()) {
                throw ConnectivityException("No network connection")
            }

            val response = httpClient.get(endpoint)
            Result.success(response)
        } catch (e: UnknownHostException) {
            Result.failure(ConnectivityException("Cannot reach server", e))
        } catch (e: SocketTimeoutException) {
            Result.failure(TimeoutException("Request timed out", e))
        } catch (e: SSLException) {
            Result.failure(SecurityException("SSL certificate error", e))
        } catch (e: IOException) {
            Result.failure(NetworkException("Network I/O error", e))
        }
    }

    private fun isNetworkAvailable(): Boolean {
        val connectivityManager = context.getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager
        val activeNetwork = connectivityManager.activeNetworkInfo
        return activeNetwork?.isConnectedOrConnecting == true
    }
}
```

#### Retrofit Error Handling

```kotlin
class ApiService {
    suspend fun getUser(userId: String): NetworkResult<User> {
        return try {
            val response = userApi.getUser(userId)
            when {
                response.isSuccessful -> {
                    response.body()?.let { user ->
                        NetworkResult.Success(user)
                    } ?: NetworkResult.Error.UnknownError(Exception("Empty response"))
                }
                response.code() == 404 -> {
                    NetworkResult.Error.NotFound("User not found")
                }
                response.code() == 401 -> {
                    NetworkResult.Error.Unauthorized("Authentication required")
                }
                response.code() in 500..599 -> {
                    NetworkResult.Error.ServerError(response.code(), "Server error")
                }
                else -> {
                    NetworkResult.Error.ClientError(response.code(), response.message())
                }
            }
        } catch (e: IOException) {
            NetworkResult.Error.NetworkError("Network connection failed")
        } catch (e: HttpException) {
            NetworkResult.Error.HttpError(e.code(), e.message())
        } catch (e: Exception) {
            NetworkResult.Error.UnknownError(e)
        }
    }
}

sealed class NetworkResult<out T> {
    data class Success<T>(val data: T) : NetworkResult<T>()

    sealed class Error : NetworkResult<Nothing>() {
        data class NetworkError(val message: String) : Error()
        data class HttpError(val code: Int, val message: String) : Error()
        data class NotFound(val message: String) : Error()
        data class Unauthorized(val message: String) : Error()
        data class ServerError(val code: Int, val message: String) : Error()
        data class ClientError(val code: Int, val message: String) : Error()
        data class UnknownError(val throwable: Throwable) : Error()
    }
}
```

### 2. Database Error Handling

Room database operations must be performed off the main thread.

#### Room DAO Error Handling

```kotlin
@Dao
interface UserDao {
    @Query("SELECT * FROM users WHERE id = :userId")
    suspend fun getUserById(userId: String): User?

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertUser(user: User): Long

    @Update
    suspend fun updateUser(user: User): Int

    @Delete
    suspend fun deleteUser(user: User): Int
}

class UserRepository(private val userDao: UserDao) {
    suspend fun getUser(userId: String): Result<User> = withContext(Dispatchers.IO) {
        try {
            val user = userDao.getUserById(userId)
                ?: return@withContext Result.failure(UserNotFoundException("User $userId not found"))

            Result.success(user)
        } catch (e: SQLiteException) {
            Result.failure(DatabaseException("Database error: ${e.message}", e))
        } catch (e: Exception) {
            Result.failure(DataException("Failed to fetch user", e))
        }
    }

    suspend fun saveUser(user: User): Result<Long> = withContext(Dispatchers.IO) {
        try {
            val id = userDao.insertUser(user)
            if (id > 0) {
                Result.success(id)
            } else {
                Result.failure(DatabaseException("Failed to insert user"))
            }
        } catch (e: SQLiteConstraintException) {
            Result.failure(ValidationException("User data violates constraints", e))
        } catch (e: SQLiteException) {
            Result.failure(DatabaseException("Database error: ${e.message}", e))
        }
    }
}
```

#### Database Migration Error Handling

```kotlin
val MIGRATION_1_2 = object : Migration(1, 2) {
    override fun migrate(database: SupportSQLiteDatabase) {
        try {
            database.execSQL("ALTER TABLE users ADD COLUMN last_login INTEGER DEFAULT 0")
        } catch (e: SQLException) {
            // Log the error and potentially trigger a fallback migration
            Log.e("Migration", "Failed to migrate database", e)
            throw MigrationException("Failed to migrate database from version 1 to 2", e)
        }
    }
}

@Database(
    entities = [User::class],
    version = 2,
    exportSchema = false
)
abstract class AppDatabase : RoomDatabase() {
    companion object {
        fun buildDatabase(context: Context): AppDatabase {
            return Room.databaseBuilder(
                context,
                AppDatabase::class.java,
                "app_database"
            )
            .addMigrations(MIGRATION_1_2)
            .fallbackToDestructiveMigration() // Use with caution
            .setQueryCallback({ sqlQuery, bindArgs ->
                Log.d("RoomQuery", "Query: $sqlQuery, Args: $bindArgs")
            }, Dispatchers.IO)
            .build()
        }
    }
}
```

### 3. Permission Error Handling

Modern Android requires runtime permission handling for sensitive operations.

#### Permission Request Pattern

```kotlin
class PermissionManager(private val activity: Activity) {
    private val requestPermissionLauncher = activity.registerForActivityResult(
        ActivityResultContracts.RequestPermission()
    ) { isGranted ->
        handlePermissionResult(isGranted)
    }

    fun requestLocationPermission(onGranted: () -> Unit, onDenied: (Boolean) -> Unit) {
        when {
            ContextCompat.checkSelfPermission(
                activity,
                Manifest.permission.ACCESS_FINE_LOCATION
            ) == PackageManager.PERMISSION_GRANTED -> {
                onGranted()
            }

            activity.shouldShowRequestPermissionRationale(
                Manifest.permission.ACCESS_FINE_LOCATION
            ) -> {
                // Show rationale and request permission
                showPermissionRationale {
                    requestPermissionLauncher.launch(Manifest.permission.ACCESS_FINE_LOCATION)
                }
            }

            else -> {
                requestPermissionLauncher.launch(Manifest.permission.ACCESS_FINE_LOCATION)
            }
        }
    }

    private fun handlePermissionResult(isGranted: Boolean) {
        if (isGranted) {
            onPermissionGranted()
        } else {
            // Check if user selected "Don't ask again"
            val shouldShowRationale = activity.shouldShowRequestPermissionRationale(
                Manifest.permission.ACCESS_FINE_LOCATION
            )
            onPermissionDenied(!shouldShowRationale)
        }
    }

    private fun showPermissionRationale(onPositive: () -> Unit) {
        AlertDialog.Builder(activity)
            .setTitle("Location Permission Required")
            .setMessage("This app needs location permission to provide location-based features.")
            .setPositiveButton("Grant") { _, _ -> onPositive() }
            .setNegativeButton("Cancel") { dialog, _ ->
                dialog.dismiss()
                onPermissionDenied(false)
            }
            .show()
    }
}
```

#### Advanced Permission Handling

```kotlin
sealed class PermissionResult {
    object Granted : PermissionResult()
    object Denied : PermissionResult()
    object PermanentlyDenied : PermissionResult()
    data class Error(val exception: Exception) : PermissionResult()
}

class AdvancedPermissionManager(private val context: Context) {
    suspend fun requestPermission(permission: String): PermissionResult {
        return suspendCoroutine { continuation ->
            when {
                hasPermission(permission) -> {
                    continuation.resume(PermissionResult.Granted)
                }

                shouldShowRationale(permission) -> {
                    // Request with rationale
                    requestWithRationale(permission) { result ->
                        continuation.resume(result)
                    }
                }

                else -> {
                    // Direct request
                    requestPermissionDirect(permission) { result ->
                        continuation.resume(result)
                    }
                }
            }
        }
    }

    private fun hasPermission(permission: String): Boolean {
        return ContextCompat.checkSelfPermission(context, permission) ==
               PackageManager.PERMISSION_GRANTED
    }

    fun handlePermissionError(result: PermissionResult, permission: String) {
        when (result) {
            is PermissionResult.Granted -> {
                // Continue with operation
            }

            is PermissionResult.Denied -> {
                showPermissionDeniedMessage()
            }

            is PermissionResult.PermanentlyDenied -> {
                showSettingsDialog(permission)
            }

            is PermissionResult.Error -> {
                Log.e("Permission", "Permission error", result.exception)
                showGenericErrorMessage()
            }
        }
    }
}
```

### 4. Lifecycle Error Handling

Android components have complex lifecycles that can cause errors if not handled properly.

#### Fragment Lifecycle Error Prevention

```kotlin
class SafeFragment : Fragment() {
    private var _binding: FragmentBinding? = null
    private val binding get() = _binding!!

    private val job = SupervisorJob()
    private val scope = CoroutineScope(Dispatchers.Main + job)

    override fun onCreateView(inflater: LayoutInflater, container: ViewGroup?, savedInstanceState: Bundle?): View? {
        _binding = FragmentBinding.inflate(inflater, container, false)
        return binding.root
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)

        // Safe lifecycle-aware operations
        viewLifecycleOwner.lifecycleScope.launch {
            try {
                loadData()
            } catch (e: Exception) {
                handleError(e)
            }
        }
    }

    private suspend fun loadData() {
        // Check if fragment is still active
        if (!isAdded || isDetached || isRemoving) {
            return
        }

        // Perform operations safely
        val data = withContext(Dispatchers.IO) {
            fetchDataFromNetwork()
        }

        // Check again before updating UI
        if (isAdded && _binding != null) {
            updateUI(data)
        }
    }

    private fun updateUI(data: Data) {
        // Safe UI updates with null checks
        _binding?.let { binding ->
            binding.textView.text = data.title
            binding.progressBar.isVisible = false
        }
    }

    override fun onDestroyView() {
        super.onDestroyView()
        job.cancel() // Cancel all coroutines
        _binding = null // Prevent memory leaks
    }
}
```

#### ViewModel Error Handling

```kotlin
class SafeViewModel : ViewModel() {
    private val _uiState = MutableStateFlow<UiState<Data>>(UiState.Loading)
    val uiState: StateFlow<UiState<Data>> = _uiState.asStateFlow()

    private val exceptionHandler = CoroutineExceptionHandler { _, exception ->
        Log.e("ViewModel", "Unhandled exception", exception)
        _uiState.value = UiState.Error("An unexpected error occurred", exception)
    }

    fun loadData() {
        viewModelScope.launch(exceptionHandler) {
            try {
                _uiState.value = UiState.Loading

                val data = repository.getData()
                _uiState.value = UiState.Success(data)

            } catch (e: CancellationException) {
                // Don't handle cancellation as error
                throw e
            } catch (e: Exception) {
                _uiState.value = UiState.Error("Failed to load data", e)
            }
        }
    }
}
```

### 5. UI Thread Error Handling

Preventing ANRs (Application Not Responding) by handling main thread violations.

```kotlin
class ThreadSafeRepository {
    private val mainDispatcher = Dispatchers.Main
    private val ioDispatcher = Dispatchers.IO

    suspend fun performNetworkOperation(): Result<Data> {
        // Ensure we're not on the main thread for network operations
        return withContext(ioDispatcher) {
            try {
                val data = networkService.fetchData()
                Result.success(data)
            } catch (e: NetworkOnMainThreadException) {
                // This should never happen due to withContext, but handle anyway
                Log.e("Repository", "Network operation attempted on main thread", e)
                Result.failure(ThreadingException("Network operation not allowed on main thread", e))
            } catch (e: Exception) {
                Result.failure(e)
            }
        }
    }

    suspend fun updateUI(data: Data) {
        // Ensure UI updates happen on main thread
        withContext(mainDispatcher) {
            try {
                updateUserInterface(data)
            } catch (e: Exception) {
                Log.e("Repository", "UI update failed", e)
            }
        }
    }
}

// Main thread checker utility
object ThreadChecker {
    fun ensureMainThread() {
        if (Looper.myLooper() != Looper.getMainLooper()) {
            throw ThreadingException("This operation must be called from the main thread")
        }
    }

    fun ensureBackgroundThread() {
        if (Looper.myLooper() == Looper.getMainLooper()) {
            throw ThreadingException("This operation cannot be called from the main thread")
        }
    }
}
```

**Key Android Error Patterns:**
- Always use background threads for network/database operations
- Handle permission denial gracefully with user education
- Implement proper lifecycle-aware error handling
- Use structured concurrency to prevent leaked operations
- Validate thread context for UI operations

---

## Coroutines Error Handling

Kotlin coroutines provide structured concurrency with sophisticated error handling mechanisms.

### 1. Structured Concurrency Fundamentals

In structured concurrency, if a coroutine encounters an exception (other than `CancellationException`), it cancels its parent with that exception.

#### Basic Exception Propagation

```kotlin
class DataService {
    suspend fun fetchDataWithStructuredConcurrency(): List<Data> = coroutineScope {
        try {
            // All child coroutines must complete successfully
            val userData = async { fetchUserData() }
            val settingsData = async { fetchSettings() }
            val preferencesData = async { fetchPreferences() }

            // If any child fails, all are cancelled and exception propagates
            listOf(userData.await(), settingsData.await(), preferencesData.await())
        } catch (e: Exception) {
            Log.e("DataService", "Failed to fetch data", e)
            throw DataException("Unable to load required data", e)
        }
    }
}
```

#### Exception Handling with supervisorScope

```kotlin
class RobustDataService {
    suspend fun fetchDataWithPartialFailure(): DataResult = supervisorScope {
        val userData = async {
            try {
                fetchUserData()
            } catch (e: Exception) {
                Log.w("DataService", "User data failed", e)
                null
            }
        }

        val settingsData = async {
            try {
                fetchSettings()
            } catch (e: Exception) {
                Log.w("DataService", "Settings failed", e)
                getDefaultSettings()
            }
        }

        // Individual failures don't cancel other operations
        DataResult(
            user = userData.await(),
            settings = settingsData.await()
        )
    }
}
```

### 2. CoroutineExceptionHandler

Used for handling uncaught exceptions in root coroutines.

#### Basic Exception Handler

```kotlin
class ExceptionHandlingService {
    private val exceptionHandler = CoroutineExceptionHandler { _, exception ->
        Log.e("CoroutineError", "Uncaught exception", exception)
        when (exception) {
            is NetworkException -> handleNetworkError(exception)
            is DatabaseException -> handleDatabaseError(exception)
            else -> handleGenericError(exception)
        }
    }

    fun startDataSync() {
        // Exception handler only works on root coroutines
        GlobalScope.launch(exceptionHandler) {
            try {
                syncData()
            } catch (e: CancellationException) {
                // Never catch CancellationException
                throw e
            } catch (e: Exception) {
                // This won't trigger the exception handler since it's caught
                Log.e("Sync", "Data sync failed", e)
            }
        }
    }

    private fun handleNetworkError(exception: NetworkException) {
        // Show user-friendly network error message
        showNotification("Network connection issue", "Check your internet connection")
    }
}
```

#### ViewModel Exception Handling

```kotlin
class UserViewModel : ViewModel() {
    private val _userState = MutableStateFlow<UiState<User>>(UiState.Loading)
    val userState: StateFlow<UiState<User>> = _userState.asStateFlow()

    private val exceptionHandler = CoroutineExceptionHandler { _, exception ->
        Log.e("UserViewModel", "Unhandled exception in ViewModel", exception)
        _userState.value = UiState.Error("An unexpected error occurred", exception)
    }

    fun loadUser(userId: String) {
        viewModelScope.launch(exceptionHandler) {
            try {
                _userState.value = UiState.Loading
                val user = userRepository.getUser(userId)
                _userState.value = UiState.Success(user)
            } catch (e: CancellationException) {
                throw e // Never catch cancellation
            } catch (e: Exception) {
                _userState.value = UiState.Error("Failed to load user", e)
            }
        }
    }
}
```

### 3. Cancellation vs Exceptions

Distinguishing between cancellation and actual errors is crucial.

```kotlin
class CancellationAwareService {
    suspend fun processData(data: List<Item>): List<ProcessedItem> {
        return data.map { item ->
            try {
                processItem(item)
            } catch (e: CancellationException) {
                // Cancellation should propagate immediately
                Log.d("Processing", "Processing cancelled")
                throw e
            } catch (e: Exception) {
                // Handle actual errors
                Log.w("Processing", "Failed to process item ${item.id}", e)
                ProcessedItem.error(item, e.message ?: "Processing failed")
            }
        }
    }

    suspend fun processWithTimeout(item: Item): ProcessedItem {
        return try {
            withTimeout(5000) {
                processItem(item)
            }
        } catch (e: TimeoutCancellationException) {
            Log.w("Processing", "Item processing timed out: ${item.id}")
            ProcessedItem.timeout(item)
        } catch (e: CancellationException) {
            // Other cancellation reasons
            throw e
        } catch (e: Exception) {
            Log.e("Processing", "Item processing failed: ${item.id}", e)
            ProcessedItem.error(item, e.message ?: "Unknown error")
        }
    }
}
```

### 4. Error Boundaries with Scopes

Creating error boundaries to contain failures within specific scopes.

```kotlin
class ErrorBoundaryService {
    // Critical operations that must not fail
    suspend fun performCriticalOperation(): Result<String> = coroutineScope {
        try {
            val result = criticalNetworkCall()
            Result.success(result)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    // Independent operations with isolated error handling
    suspend fun performParallelOperations(): ParallelResult = supervisorScope {
        val operations = listOf(
            async { performOperation("A") },
            async { performOperation("B") },
            async { performOperation("C") }
        )

        val results = operations.map { deferred ->
            try {
                deferred.await()
            } catch (e: Exception) {
                Log.w("ParallelOp", "Operation failed", e)
                null
            }
        }

        ParallelResult(
            successful = results.filterNotNull(),
            failed = results.count { it == null }
        )
    }

    // Retry with error boundary
    suspend fun performWithRetry(maxRetries: Int = 3): Result<String> {
        repeat(maxRetries) { attempt ->
            supervisorScope {
                try {
                    val result = async { riskyOperation() }
                    return@performWithRetry Result.success(result.await())
                } catch (e: CancellationException) {
                    throw e
                } catch (e: Exception) {
                    Log.w("Retry", "Attempt ${attempt + 1} failed", e)
                    if (attempt < maxRetries - 1) {
                        delay(1000 * (attempt + 1))
                    }
                }
            }
        }
        return Result.failure(Exception("Operation failed after $maxRetries attempts"))
    }
}
```

### 5. Coroutine Error Handling Best Practices

```kotlin
class BestPracticesService {
    // ✅ Good: Explicit error handling with proper cancellation support
    suspend fun goodErrorHandling(): Result<Data> {
        return try {
            val data = withContext(Dispatchers.IO) {
                fetchDataFromNetwork()
            }
            Result.success(data)
        } catch (e: CancellationException) {
            throw e // Always rethrow cancellation
        } catch (e: Exception) {
            Log.e("Service", "Data fetch failed", e)
            Result.failure(e)
        }
    }

    // ❌ Bad: runCatching breaks structured concurrency
    suspend fun badErrorHandling(): Result<Data> {
        return runCatching {
            fetchDataFromNetwork() // This catches CancellationException
        }
    }

    // ✅ Good: Proper exception handler placement
    class GoodViewModel : ViewModel() {
        private val exceptionHandler = CoroutineExceptionHandler { _, exception ->
            handleUncaughtException(exception)
        }

        fun loadData() {
            viewModelScope.launch(exceptionHandler) {
                // Exception handler works here (root coroutine)
                performDataLoading()
            }
        }
    }

    // ❌ Bad: Exception handler on child coroutines
    class BadViewModel : ViewModel() {
        private val exceptionHandler = CoroutineExceptionHandler { _, exception ->
            // This will never be called
            handleUncaughtException(exception)
        }

        fun loadData() {
            viewModelScope.launch {
                async(exceptionHandler) {
                    // Exception handler doesn't work on child coroutines
                    performDataLoading()
                }
            }
        }
    }
}
```

**Key Coroutine Error Handling Rules:**
- Never catch `CancellationException` - always rethrow it
- Use `CoroutineExceptionHandler` only on root coroutines
- Prefer explicit try-catch over `runCatching` in coroutines
- Use `supervisorScope` for independent parallel operations
- Use `coroutineScope` when all operations must succeed

---

## Reactive Error Handling

Kotlin Flow provides powerful operators for handling errors in reactive streams.

### 1. Flow Error Handling Operators

#### Basic Catch Operator

```kotlin
class NewsRepository {
    fun getLatestNews(): Flow<List<Article>> = flow {
        val articles = newsApi.getLatestArticles()
        emit(articles)
    }.catch { exception ->
        Log.e("NewsRepository", "Failed to fetch news", exception)
        when (exception) {
            is IOException -> emit(getCachedNews())
            is HttpException -> {
                if (exception.code() == 404) {
                    emit(emptyList())
                } else {
                    throw NetworkException("News service unavailable", exception)
                }
            }
            else -> throw exception
        }
    }.flowOn(Dispatchers.IO)

    private suspend fun getCachedNews(): List<Article> {
        return newsDao.getAllArticles().firstOrNull() ?: emptyList()
    }
}
```

#### Advanced Error Handling with Retry

```kotlin
class RobustApiService {
    fun fetchUserData(userId: String): Flow<User> = flow {
        val user = userApi.getUser(userId)
        emit(user)
    }.retry(3) { exception ->
        Log.w("ApiService", "Retrying due to error", exception)
        when (exception) {
            is IOException -> {
                delay(1000) // Wait before retry
                true
            }
            is HttpException -> exception.code() in 500..599 // Retry server errors
            else -> false
        }
    }.catch { exception ->
        Log.e("ApiService", "Final error after retries", exception)
        throw UserLoadException("Failed to load user after retries", exception)
    }
}
```

#### Flow Error Recovery Patterns

```kotlin
class DataSyncService {
    fun syncDataWithFallbacks(): Flow<SyncResult> = flow {
        emit(SyncResult.InProgress)

        // Try primary source
        val primaryData = primaryDataSource.getData()
        emit(SyncResult.Success(primaryData, "primary"))

    }.catch { primaryError ->
        Log.w("Sync", "Primary source failed, trying secondary", primaryError)

        // Emit intermediate state
        emit(SyncResult.Fallback("Primary source unavailable"))

        // Try secondary source
        secondaryDataSource.getData()
            .catch { secondaryError ->
                Log.e("Sync", "Secondary source also failed", secondaryError)
                emit(SyncResult.Error("All data sources failed"))
            }
            .collect { secondaryData ->
                emit(SyncResult.Success(secondaryData, "secondary"))
            }
    }

    sealed class SyncResult {
        object InProgress : SyncResult()
        data class Success(val data: Data, val source: String) : SyncResult()
        data class Fallback(val message: String) : SyncResult()
        data class Error(val message: String) : SyncResult()
    }
}
```

### 2. StateFlow Error Handling

StateFlow requires different error handling patterns since it never completes.

```kotlin
class UserStateManager {
    private val _userState = MutableStateFlow<UserState>(UserState.Loading)
    val userState: StateFlow<UserState> = _userState.asStateFlow()

    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())

    fun loadUser(userId: String) {
        scope.launch {
            _userState.value = UserState.Loading

            try {
                val user = userRepository.getUser(userId)
                _userState.value = UserState.Success(user)
            } catch (e: CancellationException) {
                throw e
            } catch (e: Exception) {
                Log.e("UserState", "Failed to load user", e)
                _userState.value = UserState.Error(e.message ?: "Unknown error")
            }
        }
    }

    fun refreshUser() {
        if (_userState.value is UserState.Success) {
            val currentUser = (_userState.value as UserState.Success).user
            loadUser(currentUser.id)
        }
    }

    sealed class UserState {
        object Loading : UserState()
        data class Success(val user: User) : UserState()
        data class Error(val message: String) : UserState()
    }
}
```

### 3. SharedFlow Error Handling

SharedFlow error handling is more complex due to its hot nature.

```kotlin
class EventBus {
    private val _events = MutableSharedFlow<Event>()
    val events: SharedFlow<Event> = _events.asSharedFlow()

    private val _errors = MutableSharedFlow<ErrorEvent>()
    val errors: SharedFlow<ErrorEvent> = _errors.asSharedFlow()

    fun emitEvent(event: Event) {
        try {
            if (!_events.tryEmit(event)) {
                Log.w("EventBus", "Failed to emit event, buffer full")
            }
        } catch (e: Exception) {
            Log.e("EventBus", "Error emitting event", e)
            _errors.tryEmit(ErrorEvent("Failed to emit event", e))
        }
    }

    // Transform SharedFlow for error handling
    fun safeEvents(): Flow<Result<Event>> = events.map { event ->
        try {
            Result.success(validateEvent(event))
        } catch (e: Exception) {
            Result.failure(e)
        }
    }.catch { exception ->
        emit(Result.failure(exception))
    }

    data class ErrorEvent(val message: String, val exception: Exception)
}
```

### 4. Combining Multiple Flows with Error Handling

```kotlin
class CombinedDataService {
    fun getCombinedUserData(userId: String): Flow<CombinedResult> =
        combine(
            getUserProfile(userId).catch { emit(null) },
            getUserSettings(userId).catch { emit(UserSettings.default()) },
            getUserPreferences(userId).catch { emit(UserPreferences.default()) }
        ) { profile, settings, preferences ->
            when {
                profile == null -> CombinedResult.Error("Failed to load user profile")
                else -> CombinedResult.Success(
                    CombinedUserData(profile, settings, preferences)
                )
            }
        }.catch { exception ->
            Log.e("CombinedData", "Error combining data", exception)
            emit(CombinedResult.Error("Failed to load user data"))
        }

    private fun getUserProfile(userId: String): Flow<UserProfile?> = flow {
        val profile = userApi.getProfile(userId)
        emit(profile)
    }.retry(2) { exception ->
        exception is IOException
    }

    sealed class CombinedResult {
        data class Success(val data: CombinedUserData) : CombinedResult()
        data class Error(val message: String) : CombinedResult()
    }
}
```

### 5. Flow Error Handling in UI

```kotlin
@Composable
fun UserScreen(userId: String) {
    val userRepository = hiltViewModel<UserRepository>()

    val userState by userRepository.getUserFlow(userId)
        .catch { exception ->
            // Handle Flow errors in UI
            Log.e("UserScreen", "User flow error", exception)
            emit(UserUiState.Error("Failed to load user"))
        }
        .collectAsState(UserUiState.Loading)

    when (userState) {
        is UserUiState.Loading -> LoadingIndicator()
        is UserUiState.Success -> UserContent(userState.user)
        is UserUiState.Error -> ErrorMessage(
            message = userState.message,
            onRetry = { userRepository.retryLoadUser(userId) }
        )
    }
}

// Repository with Flow error handling
class UserRepository {
    private val refreshTrigger = MutableSharedFlow<String>()

    fun getUserFlow(userId: String): Flow<UserUiState> =
        merge(
            flowOf(userId),
            refreshTrigger.filter { it == userId }
        )
        .flatMapLatest { id ->
            flow {
                emit(UserUiState.Loading)
                val user = userApi.getUser(id)
                emit(UserUiState.Success(user))
            }
        }
        .catch { exception ->
            when (exception) {
                is IOException -> emit(UserUiState.Error("Network error"))
                is HttpException -> when (exception.code()) {
                    404 -> emit(UserUiState.Error("User not found"))
                    else -> emit(UserUiState.Error("Server error"))
                }
                else -> emit(UserUiState.Error("Unknown error"))
            }
        }
        .flowOn(Dispatchers.IO)

    fun retryLoadUser(userId: String) {
        refreshTrigger.tryEmit(userId)
    }
}
```

**Flow Error Handling Best Practices:**
- Use `catch` for graceful error recovery
- Combine `retry` with exponential backoff for transient errors
- Handle SharedFlow errors by transforming to cold flows
- Emit error states instead of throwing exceptions in UI flows
- Use `flowOn` to handle errors on appropriate dispatchers

---

## Functional Error Handling

Functional programming approaches to error handling provide composable and predictable error management patterns.

### 1. Arrow Either Type

Arrow's `Either` type represents values that can be either a success (Right) or a failure (Left).

#### Basic Either Usage

```kotlin
import arrow.core.Either
import arrow.core.left
import arrow.core.right

sealed class UserError {
    object NotFound : UserError()
    object NetworkError : UserError()
    data class ValidationError(val message: String) : UserError()
}

class FunctionalUserRepository {
    suspend fun getUser(userId: String): Either<UserError, User> {
        return try {
            if (userId.isEmpty()) {
                UserError.ValidationError("User ID cannot be empty").left()
            } else {
                val user = userApi.getUser(userId)
                user.right()
            }
        } catch (e: IOException) {
            UserError.NetworkError.left()
        } catch (e: HttpException) {
            when (e.code()) {
                404 -> UserError.NotFound.left()
                else -> UserError.NetworkError.left()
            }
        }
    }

    suspend fun getUserWithFallback(userId: String): Either<UserError, User> {
        return getUser(userId).fold(
            { error ->
                when (error) {
                    is UserError.NetworkError -> getCachedUser(userId)
                    else -> error.left()
                }
            },
            { user -> user.right() }
        )
    }
}
```

#### Chaining Either Operations

```kotlin
class UserService {
    suspend fun processUserRegistration(
        email: String,
        password: String,
        name: String
    ): Either<RegistrationError, User> {
        return validateEmail(email)
            .flatMap { validatePassword(password) }
            .flatMap { validateName(name) }
            .flatMap { createUser(email, password, name) }
            .flatMap { user -> sendWelcomeEmail(user).map { user } }
    }

    private fun validateEmail(email: String): Either<RegistrationError, String> {
        return if (android.util.Patterns.EMAIL_ADDRESS.matcher(email).matches()) {
            email.right()
        } else {
            RegistrationError.InvalidEmail.left()
        }
    }

    private fun validatePassword(password: String): Either<RegistrationError, String> {
        return when {
            password.length < 8 -> RegistrationError.WeakPassword("Password too short").left()
            !password.any { it.isDigit() } -> RegistrationError.WeakPassword("Password needs digits").left()
            else -> password.right()
        }
    }

    private suspend fun createUser(email: String, password: String, name: String): Either<RegistrationError, User> {
        return try {
            val user = userRepository.createUser(email, password, name)
            user.right()
        } catch (e: DuplicateEmailException) {
            RegistrationError.EmailAlreadyExists.left()
        } catch (e: Exception) {
            RegistrationError.DatabaseError.left()
        }
    }

    sealed class RegistrationError {
        object InvalidEmail : RegistrationError()
        data class WeakPassword(val reason: String) : RegistrationError()
        object EmailAlreadyExists : RegistrationError()
        object DatabaseError : RegistrationError()
    }
}
```

#### Error Accumulation with Validated

```kotlin
import arrow.core.Validated
import arrow.core.invalid
import arrow.core.valid
import arrow.core.zip

data class UserRegistrationForm(
    val email: String,
    val password: String,
    val confirmPassword: String,
    val age: Int,
    val terms: Boolean
)

class FormValidator {
    fun validateRegistrationForm(form: UserRegistrationForm): Validated<List<ValidationError>, ValidatedForm> {
        return validateEmail(form.email)
            .zip(
                validatePassword(form.password),
                validatePasswordMatch(form.password, form.confirmPassword),
                validateAge(form.age),
                validateTerms(form.terms)
            ) { email, password, _, age, _ ->
                ValidatedForm(email, password, age)
            }
    }

    private fun validateEmail(email: String): Validated<List<ValidationError>, String> {
        return if (android.util.Patterns.EMAIL_ADDRESS.matcher(email).matches()) {
            email.valid()
        } else {
            listOf(ValidationError.InvalidEmail).invalid()
        }
    }

    private fun validatePassword(password: String): Validated<List<ValidationError>, String> {
        val errors = mutableListOf<ValidationError>()

        if (password.length < 8) {
            errors.add(ValidationError.PasswordTooShort)
        }
        if (!password.any { it.isDigit() }) {
            errors.add(ValidationError.PasswordNeedsDigit)
        }
        if (!password.any { it.isUpperCase() }) {
            errors.add(ValidationError.PasswordNeedsUppercase)
        }

        return if (errors.isEmpty()) {
            password.valid()
        } else {
            errors.invalid()
        }
    }

    sealed class ValidationError {
        object InvalidEmail : ValidationError()
        object PasswordTooShort : ValidationError()
        object PasswordNeedsDigit : ValidationError()
        object PasswordNeedsUppercase : ValidationError()
        object PasswordMismatch : ValidationError()
        object AgeTooYoung : ValidationError()
        object TermsNotAccepted : ValidationError()
    }
}
```

### 2. Railway-Oriented Programming

Railway-oriented programming creates a "happy path" and "error path" that operations can switch between.

```kotlin
// Extension functions for railway-oriented programming
inline fun <A, B, C> Either<A, B>.bind(f: (B) -> Either<A, C>): Either<A, C> = flatMap(f)

inline fun <A, B> Either<A, B>.tee(f: (B) -> Unit): Either<A, B> =
    map { value -> f(value); value }

class OrderProcessor {
    suspend fun processOrder(orderRequest: OrderRequest): Either<OrderError, Order> {
        return validateOrder(orderRequest)
            .bind { order -> checkInventory(order) }
            .bind { order -> calculatePricing(order) }
            .bind { order -> processPayment(order) }
            .bind { order -> createShipment(order) }
            .tee { order -> sendConfirmationEmail(order) }
            .tee { order -> logOrderProcessed(order) }
    }

    private fun validateOrder(request: OrderRequest): Either<OrderError, Order> {
        return when {
            request.items.isEmpty() -> OrderError.EmptyOrder.left()
            request.customerId.isEmpty() -> OrderError.InvalidCustomer.left()
            else -> Order.fromRequest(request).right()
        }
    }

    private suspend fun checkInventory(order: Order): Either<OrderError, Order> {
        return try {
            val availability = inventoryService.checkAvailability(order.items)
            if (availability.allAvailable) {
                order.right()
            } else {
                OrderError.InsufficientInventory(availability.unavailableItems).left()
            }
        } catch (e: Exception) {
            OrderError.InventoryCheckFailed.left()
        }
    }

    private suspend fun processPayment(order: Order): Either<OrderError, Order> {
        return try {
            val paymentResult = paymentService.processPayment(order.paymentInfo, order.total)
            if (paymentResult.isSuccessful) {
                order.copy(paymentId = paymentResult.transactionId).right()
            } else {
                OrderError.PaymentFailed(paymentResult.errorMessage).left()
            }
        } catch (e: Exception) {
            OrderError.PaymentProcessingError.left()
        }
    }

    sealed class OrderError {
        object EmptyOrder : OrderError()
        object InvalidCustomer : OrderError()
        data class InsufficientInventory(val items: List<String>) : OrderError()
        object InventoryCheckFailed : OrderError()
        data class PaymentFailed(val reason: String) : OrderError()
        object PaymentProcessingError : OrderError()
        object ShipmentCreationFailed : OrderError()
    }
}
```

### 3. Functional Error Handling in Android

#### Repository with Either

```kotlin
class FunctionalUserRepository(
    private val userApi: UserApi,
    private val userDao: UserDao
) {
    suspend fun getUserProfile(userId: String): Either<UserError, UserProfile> {
        return getFromCache(userId)
            .flatMap { cachedUser ->
                if (cachedUser.isExpired()) {
                    getFromNetwork(userId)
                        .map { networkUser ->
                            cacheUser(networkUser)
                            networkUser
                        }
                } else {
                    cachedUser.right()
                }
            }
    }

    private suspend fun getFromNetwork(userId: String): Either<UserError, UserProfile> {
        return Either.catch {
            userApi.getUserProfile(userId)
        }.mapLeft { exception ->
            when (exception) {
                is IOException -> UserError.NetworkError
                is HttpException -> when (exception.code()) {
                    404 -> UserError.NotFound
                    401 -> UserError.Unauthorized
                    else -> UserError.ServerError
                }
                else -> UserError.UnknownError
            }
        }
    }

    private suspend fun getFromCache(userId: String): Either<UserError, UserProfile> {
        return Either.catch {
            userDao.getUserProfile(userId) ?: throw NoSuchElementException()
        }.mapLeft {
            UserError.NotInCache
        }
    }
}
```

#### ViewModel with Either

```kotlin
class FunctionalUserViewModel : ViewModel() {
    private val _userState = MutableStateFlow<UserViewState>(UserViewState.Loading)
    val userState: StateFlow<UserViewState> = _userState.asStateFlow()

    fun loadUser(userId: String) {
        viewModelScope.launch {
            userRepository.getUserProfile(userId)
                .fold(
                    { error -> _userState.value = error.toViewState() },
                    { user -> _userState.value = UserViewState.Success(user) }
                )
        }
    }

    private fun UserError.toViewState(): UserViewState.Error {
        return when (this) {
            UserError.NotFound -> UserViewState.Error("User not found")
            UserError.NetworkError -> UserViewState.Error("Check your internet connection")
            UserError.Unauthorized -> UserViewState.Error("Please log in again")
            UserError.ServerError -> UserViewState.Error("Server is temporarily unavailable")
            UserError.UnknownError -> UserViewState.Error("Something went wrong")
            UserError.NotInCache -> UserViewState.Error("User data not available")
        }
    }

    sealed class UserViewState {
        object Loading : UserViewState()
        data class Success(val user: UserProfile) : UserViewState()
        data class Error(val message: String) : UserViewState()
    }
}
```

### 4. Composing Either with Coroutines

```kotlin
class AsyncEitherService {
    suspend fun processMultipleUsers(userIds: List<String>): Either<BatchError, List<User>> {
        return Either.catch {
            userIds.map { userId ->
                async { userRepository.getUser(userId) }
            }.awaitAll()
        }.flatMap { results ->
            val errors = results.mapNotNull { it.leftOrNull() }
            val users = results.mapNotNull { it.orNull() }

            when {
                errors.isNotEmpty() -> BatchError.PartialFailure(errors).left()
                users.size != userIds.size -> BatchError.MissingUsers.left()
                else -> users.right()
            }
        }.mapLeft { exception ->
            BatchError.ProcessingError(exception.message ?: "Unknown error")
        }
    }

    suspend fun parallelDataFetch(userId: String): Either<DataError, CompleteUserData> {
        return supervisorScope {
            val userDeferred = async { userRepository.getUser(userId) }
            val settingsDeferred = async { settingsRepository.getUserSettings(userId) }
            val preferencesDeferred = async { preferencesRepository.getUserPreferences(userId) }

            val user = userDeferred.await()
            val settings = settingsDeferred.await()
            val preferences = preferencesDeferred.await()

            user.flatMap { u ->
                settings.flatMap { s ->
                    preferences.map { p ->
                        CompleteUserData(u, s, p)
                    }
                }
            }.mapLeft { error ->
                DataError.CombinationFailed(error.toString())
            }
        }
    }

    sealed class BatchError {
        data class PartialFailure(val errors: List<UserError>) : BatchError()
        object MissingUsers : BatchError()
        data class ProcessingError(val message: String) : BatchError()
    }
}
```

**Functional Error Handling Benefits:**
- **Composable**: Operations chain together cleanly
- **Explicit**: Errors are part of the type system
- **Predictable**: No hidden exceptions in the happy path
- **Testable**: Easy to test both success and error paths
- **Maintainable**: Clear separation between business logic and error handling

---

## Error Recovery Strategies

Robust mobile applications require sophisticated error recovery mechanisms to handle transient failures and provide good user experience.

### 1. Retry Patterns

#### Basic Retry Implementation

```kotlin
class RetryService {
    suspend fun <T> retryOperation(
        maxAttempts: Int = 3,
        delayMs: Long = 1000,
        operation: suspend () -> T
    ): T {
        repeat(maxAttempts - 1) { attempt ->
            try {
                return operation()
            } catch (e: Exception) {
                Log.w("Retry", "Attempt ${attempt + 1} failed", e)
                delay(delayMs)
            }
        }
        // Final attempt without catching exception
        return operation()
    }

    suspend fun fetchDataWithRetry(): User {
        return retryOperation(maxAttempts = 3) {
            userApi.getCurrentUser()
        }
    }
}
```

#### Exponential Backoff with Jitter

```kotlin
class ExponentialBackoffRetry {
    suspend fun <T> retryWithExponentialBackoff(
        maxAttempts: Int = 3,
        baseDelayMs: Long = 1000,
        maxDelayMs: Long = 30000,
        backoffMultiplier: Double = 2.0,
        jitterFactor: Double = 0.1,
        operation: suspend () -> T
    ): Result<T> {
        var lastException: Exception? = null

        repeat(maxAttempts) { attempt ->
            try {
                return Result.success(operation())
            } catch (e: Exception) {
                lastException = e

                if (attempt < maxAttempts - 1) {
                    val delay = calculateDelay(attempt, baseDelayMs, backoffMultiplier, maxDelayMs, jitterFactor)
                    Log.w("Retry", "Attempt ${attempt + 1} failed, retrying in ${delay}ms", e)
                    delay(delay)
                }
            }
        }

        return Result.failure(lastException ?: Exception("Max retry attempts exceeded"))
    }

    private fun calculateDelay(
        attempt: Int,
        baseDelayMs: Long,
        backoffMultiplier: Double,
        maxDelayMs: Long,
        jitterFactor: Double
    ): Long {
        val exponentialDelay = (baseDelayMs * backoffMultiplier.pow(attempt)).toLong()
        val cappedDelay = minOf(exponentialDelay, maxDelayMs)

        // Add jitter to prevent thundering herd
        val jitter = (cappedDelay * jitterFactor * Random.nextDouble()).toLong()
        return cappedDelay + jitter
    }
}

// Usage example
class NewsRepository {
    suspend fun getLatestNews(): Result<List<Article>> {
        return ExponentialBackoffRetry().retryWithExponentialBackoff(
            maxAttempts = 3,
            baseDelayMs = 500
        ) {
            newsApi.getLatestArticles()
        }
    }
}
```

#### Conditional Retry Strategy

```kotlin
class SmartRetryService {
    suspend fun <T> retryWithCondition(
        maxAttempts: Int = 3,
        shouldRetry: (Exception) -> Boolean,
        operation: suspend () -> T
    ): Result<T> {
        var lastException: Exception? = null

        repeat(maxAttempts) { attempt ->
            try {
                return Result.success(operation())
            } catch (e: Exception) {
                lastException = e

                if (attempt < maxAttempts - 1 && shouldRetry(e)) {
                    val delay = (1000 * (attempt + 1)).toLong()
                    Log.w("SmartRetry", "Retryable error on attempt ${attempt + 1}", e)
                    delay(delay)
                } else {
                    Log.e("SmartRetry", "Non-retryable error or max attempts reached", e)
                    break
                }
            }
        }

        return Result.failure(lastException ?: Exception("Operation failed"))
    }

    fun createRetryPredicate(): (Exception) -> Boolean = { exception ->
        when (exception) {
            is SocketTimeoutException -> true
            is UnknownHostException -> true
            is ConnectException -> true
            is HttpException -> exception.code() in listOf(408, 429, 502, 503, 504)
            else -> false
        }
    }
}

// Mobile-specific retry example
class LocationService {
    suspend fun getCurrentLocation(): Result<Location> {
        return SmartRetryService().retryWithCondition(
            maxAttempts = 3,
            shouldRetry = { exception ->
                when (exception) {
                    is SecurityException -> false // Permission denied
                    is LocationSettingsException -> false // Settings need user intervention
                    is LocationTimeoutException -> true // Temporary GPS issue
                    else -> true
                }
            }
        ) {
            locationProvider.getCurrentLocation()
        }
    }
}
```

### 2. Circuit Breaker Pattern

```kotlin
class CircuitBreaker(
    private val failureThreshold: Int = 5,
    private val recoveryTimeoutMs: Long = 30000,
    private val successThreshold: Int = 2
) {
    private var state: CircuitBreakerState = CircuitBreakerState.Closed
    private var failureCount = 0
    private var lastFailureTime = 0L
    private var successCount = 0

    suspend fun <T> execute(operation: suspend () -> T): Result<T> {
        return when (state) {
            CircuitBreakerState.Closed -> executeClosed(operation)
            CircuitBreakerState.Open -> executeOpen(operation)
            CircuitBreakerState.HalfOpen -> executeHalfOpen(operation)
        }
    }

    private suspend fun <T> executeClosed(operation: suspend () -> T): Result<T> {
        return try {
            val result = operation()
            onSuccess()
            Result.success(result)
        } catch (e: Exception) {
            onFailure()
            Result.failure(e)
        }
    }

    private fun <T> executeOpen(operation: suspend () -> T): Result<T> {
        if (System.currentTimeMillis() - lastFailureTime >= recoveryTimeoutMs) {
            state = CircuitBreakerState.HalfOpen
            successCount = 0
            Log.i("CircuitBreaker", "Transitioning to HalfOpen state")
            return executeHalfOpen(operation)
        }

        return Result.failure(CircuitBreakerException("Circuit breaker is OPEN"))
    }

    private suspend fun <T> executeHalfOpen(operation: suspend () -> T): Result<T> {
        return try {
            val result = operation()
            onSuccessInHalfOpen()
            Result.success(result)
        } catch (e: Exception) {
            onFailureInHalfOpen()
            Result.failure(e)
        }
    }

    private fun onSuccess() {
        failureCount = 0
    }

    private fun onFailure() {
        failureCount++
        lastFailureTime = System.currentTimeMillis()
        if (failureCount >= failureThreshold) {
            state = CircuitBreakerState.Open
            Log.w("CircuitBreaker", "Circuit breaker OPENED after $failureCount failures")
        }
    }

    private fun onSuccessInHalfOpen() {
        successCount++
        if (successCount >= successThreshold) {
            state = CircuitBreakerState.Closed
            failureCount = 0
            Log.i("CircuitBreaker", "Circuit breaker CLOSED after $successCount successes")
        }
    }

    private fun onFailureInHalfOpen() {
        state = CircuitBreakerState.Open
        failureCount++
        lastFailureTime = System.currentTimeMillis()
        Log.w("CircuitBreaker", "Circuit breaker returned to OPEN state")
    }

    enum class CircuitBreakerState {
        Closed, Open, HalfOpen
    }

    class CircuitBreakerException(message: String) : Exception(message)
}

// Usage in repository
class ResilientApiRepository {
    private val circuitBreaker = CircuitBreaker()

    suspend fun fetchUserData(userId: String): Result<User> {
        return circuitBreaker.execute {
            userApi.getUser(userId)
        }
    }
}
```

### 3. Fallback Mechanisms

#### Cache Fallback Strategy

```kotlin
class CacheFallbackService {
    private val memoryCache = LruCache<String, User>(100)

    suspend fun getUserWithFallbacks(userId: String): Result<User> {
        // Try network first
        return fetchFromNetwork(userId)
            .recoverCatching { networkError ->
                Log.w("UserService", "Network failed, trying cache", networkError)

                // Try persistent cache
                fetchFromPersistentCache(userId).getOrNull()
                    ?: fetchFromMemoryCache(userId).getOrNull()
                    ?: getDefaultUser(userId).also {
                        Log.i("UserService", "Using default user data")
                    }
            }
    }

    private suspend fun fetchFromNetwork(userId: String): Result<User> {
        return try {
            val user = userApi.getUser(userId)
            // Cache successful network result
            cacheUser(user)
            Result.success(user)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    private suspend fun fetchFromPersistentCache(userId: String): Result<User> {
        return try {
            val user = userDao.getUser(userId)
            if (user != null && !user.isExpired()) {
                memoryCache.put(userId, user)
                Result.success(user)
            } else {
                Result.failure(CacheException("User not in cache or expired"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    private fun fetchFromMemoryCache(userId: String): Result<User> {
        return memoryCache.get(userId)?.let { user ->
            if (!user.isExpired()) {
                Result.success(user)
            } else {
                Result.failure(CacheException("Memory cache expired"))
            }
        } ?: Result.failure(CacheException("Not in memory cache"))
    }

    private fun getDefaultUser(userId: String): User {
        return User(
            id = userId,
            name = "Guest User",
            email = "",
            isDefault = true
        )
    }
}
```

#### Progressive Enhancement Fallback

```kotlin
class ProgressiveEnhancementService {
    suspend fun getEnhancedUserData(userId: String): EnhancedUserData {
        val coreData = getCoreUserData(userId)

        return EnhancedUserData(
            core = coreData,
            profile = getUserProfile(userId).getOrNull(),
            preferences = getUserPreferences(userId).getOrNull(),
            socialData = getSocialData(userId).getOrNull(),
            analytics = getAnalyticsData(userId).getOrNull()
        )
    }

    private suspend fun getCoreUserData(userId: String): CoreUserData {
        // This must succeed - use fallbacks
        return fetchCoreData(userId).getOrElse {
            Log.w("UserService", "Using fallback core data", it)
            getDefaultCoreData(userId)
        }
    }

    private suspend fun getUserProfile(userId: String): Result<UserProfile> {
        return try {
            Result.success(profileApi.getUserProfile(userId))
        } catch (e: Exception) {
            Log.w("UserService", "Profile unavailable", e)
            Result.failure(e)
        }
    }

    // Each enhancement fails independently
    private suspend fun getUserPreferences(userId: String): Result<UserPreferences> {
        return circuitBreaker.execute {
            preferencesApi.getUserPreferences(userId)
        }
    }
}
```

### 4. Offline Mode Support

```kotlin
class OfflineModeService {
    private val connectivityManager: ConnectivityManager

    suspend fun syncDataWithOfflineSupport(): SyncResult {
        return if (isOnline()) {
            performOnlineSync()
        } else {
            performOfflineOperations()
        }
    }

    private suspend fun performOnlineSync(): SyncResult {
        return try {
            // Upload pending offline changes
            uploadPendingChanges()

            // Download latest data
            val latestData = downloadLatestData()

            // Clear offline queue
            clearOfflineQueue()

            SyncResult.Success(latestData)
        } catch (e: Exception) {
            Log.w("Sync", "Online sync failed, switching to offline mode", e)
            performOfflineOperations()
        }
    }

    private suspend fun performOfflineOperations(): SyncResult {
        // Work with cached data
        val cachedData = getCachedData()

        // Queue operations for later sync
        queuePendingOperations()

        return SyncResult.Offline(cachedData, getPendingOperationsCount())
    }

    private fun isOnline(): Boolean {
        val activeNetwork = connectivityManager.activeNetworkInfo
        return activeNetwork?.isConnectedOrConnecting == true
    }

    sealed class SyncResult {
        data class Success(val data: Data) : SyncResult()
        data class Offline(val cachedData: Data, val pendingOperations: Int) : SyncResult()
        data class Error(val message: String, val cachedData: Data?) : SyncResult()
    }
}
```

### 5. Error Recovery in UI

```kotlin
@Composable
fun ErrorRecoveryScreen(
    uiState: UiState<Data>,
    onRetry: () -> Unit,
    onRefresh: () -> Unit
) {
    when (uiState) {
        is UiState.Loading -> LoadingIndicator()

        is UiState.Success -> {
            DataContent(data = uiState.data, onRefresh = onRefresh)
        }

        is UiState.Error -> {
            ErrorRecoveryContent(
                error = uiState,
                onRetry = onRetry,
                onRefresh = onRefresh
            )
        }
    }
}

@Composable
fun ErrorRecoveryContent(
    error: UiState.Error,
    onRetry: () -> Unit,
    onRefresh: () -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        val (title, description, primaryAction, secondaryAction) = when (error.type) {
            ErrorType.Network -> ErrorRecoveryStrings(
                title = "Connection Problem",
                description = "Check your internet connection and try again",
                primaryAction = "Retry",
                secondaryAction = "Use Offline Mode"
            )

            ErrorType.Server -> ErrorRecoveryStrings(
                title = "Server Error",
                description = "Our servers are experiencing issues",
                primaryAction = "Try Again",
                secondaryAction = "Check Status"
            )

            ErrorType.Authentication -> ErrorRecoveryStrings(
                title = "Authentication Required",
                description = "Please sign in to continue",
                primaryAction = "Sign In",
                secondaryAction = "Use as Guest"
            )
        }

        Text(
            text = title,
            style = MaterialTheme.typography.h6
        )

        Text(
            text = description,
            style = MaterialTheme.typography.body2,
            textAlign = TextAlign.Center
        )

        Spacer(modifier = Modifier.height(16.dp))

        Button(onClick = onRetry) {
            Text(primaryAction)
        }

        TextButton(onClick = onRefresh) {
            Text(secondaryAction)
        }
    }
}
```

**Error Recovery Best Practices:**
- Implement exponential backoff for transient failures
- Use circuit breakers for dependent service failures
- Provide multiple fallback layers (cache, default values)
- Support offline functionality where possible
- Design user-friendly error recovery interfaces
- Monitor and log recovery attempts for analysis

---

## Testing Error Handling

Comprehensive error handling testing ensures your mobile application behaves correctly under various failure scenarios.

### 1. Unit Testing Error Scenarios

#### Testing with JUnit and MockK

```kotlin
class UserRepositoryTest {

    @MockK
    private lateinit var userApi: UserApi

    @MockK
    private lateinit var userDao: UserDao

    private lateinit var userRepository: UserRepository

    @BeforeEach
    fun setup() {
        MockKAnnotations.init(this)
        userRepository = UserRepository(userApi, userDao)
    }

    @Test
    fun `getUser returns success when api call succeeds`() = runTest {
        // Given
        val userId = "123"
        val expectedUser = User(userId, "John Doe", "john@example.com")
        coEvery { userApi.getUser(userId) } returns expectedUser

        // When
        val result = userRepository.getUser(userId)

        // Then
        assertThat(result.isSuccess).isTrue()
        assertThat(result.getOrNull()).isEqualTo(expectedUser)
    }

    @Test
    fun `getUser returns failure when network error occurs`() = runTest {
        // Given
        val userId = "123"
        val networkException = IOException("Network error")
        coEvery { userApi.getUser(userId) } throws networkException

        // When
        val result = userRepository.getUser(userId)

        // Then
        assertThat(result.isFailure).isTrue()
        assertThat(result.exceptionOrNull()).isInstanceOf(NetworkException::class.java)
        assertThat(result.exceptionOrNull()?.cause).isEqualTo(networkException)
    }

    @Test
    fun `getUser returns user not found when 404 error occurs`() = runTest {
        // Given
        val userId = "123"
        val httpException = HttpException(Response.error<Any>(404, "Not Found".toResponseBody()))
        coEvery { userApi.getUser(userId) } throws httpException

        // When
        val result = userRepository.getUser(userId)

        // Then
        assertThat(result.isFailure).isTrue()
        val exception = result.exceptionOrNull() as? UserNotFoundException
        assertThat(exception?.message).contains(userId)
    }

    @Test
    fun `getUser falls back to cache when network fails`() = runTest {
        // Given
        val userId = "123"
        val cachedUser = User(userId, "Cached User", "cached@example.com")
        coEvery { userApi.getUser(userId) } throws IOException("Network error")
        coEvery { userDao.getUser(userId) } returns cachedUser

        // When
        val result = userRepository.getUserWithFallback(userId)

        // Then
        assertThat(result.isSuccess).isTrue()
        assertThat(result.getOrNull()).isEqualTo(cachedUser)
    }
}
```

#### Testing Coroutine Exception Handling

```kotlin
class CoroutineErrorHandlingTest {

    @Test
    fun `viewModel handles coroutine exceptions properly`() = runTest {
        // Given
        val mockRepository = mockk<UserRepository>()
        val exception = RuntimeException("Test exception")
        coEvery { mockRepository.getUser(any()) } throws exception

        val viewModel = UserViewModel(mockRepository)

        // When
        viewModel.loadUser("123")
        advanceUntilIdle() // Wait for coroutine completion

        // Then
        val state = viewModel.userState.value
        assertThat(state).isInstanceOf(UiState.Error::class.java)
        val errorState = state as UiState.Error
        assertThat(errorState.message).isEqualTo("Failed to load user")
    }

    @Test
    fun `structured concurrency cancels children on exception`() = runTest {
        // Given
        val longRunningTask = mockk<LongRunningTask>()
        coEvery { longRunningTask.execute() } coAnswers {
            delay(10000) // Long running task
            "Success"
        }

        val failingTask = mockk<FailingTask>()
        coEvery { failingTask.execute() } throws RuntimeException("Immediate failure")

        val service = ConcurrentService(longRunningTask, failingTask)

        // When & Then
        assertThrows<RuntimeException> {
            runBlocking {
                service.executeParallelTasks()
            }
        }

        // Verify the long-running task was cancelled
        coVerify(exactly = 1) { longRunningTask.execute() }
        // The long-running task should not complete
    }

    @Test
    fun `supervisorScope allows partial failures`() = runTest {
        // Given
        val mockRepository = mockk<DataRepository>()
        coEvery { mockRepository.getUserData(any()) } returns "user data"
        coEvery { mockRepository.getSettingsData(any()) } throws RuntimeException("Settings failed")
        coEvery { mockRepository.getPreferencesData(any()) } returns "preferences data"

        val service = SupervisorScopeService(mockRepository)

        // When
        val result = service.loadAllDataWithPartialFailure("123")

        // Then
        assertThat(result.userData).isEqualTo("user data")
        assertThat(result.settingsData).isNull()
        assertThat(result.preferencesData).isEqualTo("preferences data")
    }
}
```

### 2. Integration Testing Error Scenarios

#### Testing with Robolectric

```kotlin
@RunWith(RobolectricTestRunner::class)
@Config(sdk = [Build.VERSION_CODES.P])
class UserFragmentIntegrationTest {

    private lateinit var scenario: FragmentScenario<UserFragment>
    private lateinit var mockRepository: UserRepository

    @Before
    fun setup() {
        mockRepository = mockk<UserRepository>()

        // Inject mock repository
        every { ServiceLocator.userRepository } returns mockRepository

        scenario = launchFragmentInContainer<UserFragment>()
    }

    @Test
    fun `fragment shows error message when user loading fails`() {
        // Given
        val errorMessage = "Failed to load user"
        coEvery { mockRepository.getUser(any()) } returns Result.failure(
            Exception(errorMessage)
        )

        // When
        scenario.onFragment { fragment ->
            fragment.loadUser("123")
        }

        // Then
        onView(withId(R.id.error_message))
            .check(matches(isDisplayed()))
            .check(matches(withText(containsString(errorMessage))))

        onView(withId(R.id.retry_button))
            .check(matches(isDisplayed()))
    }

    @Test
    fun `fragment retries loading when retry button clicked`() {
        // Given
        coEvery { mockRepository.getUser(any()) } returns Result.failure(Exception("Error"))

        scenario.onFragment { fragment ->
            fragment.loadUser("123")
        }

        // Reset mock to return success
        val user = User("123", "John Doe", "john@example.com")
        coEvery { mockRepository.getUser(any()) } returns Result.success(user)

        // When
        onView(withId(R.id.retry_button))
            .perform(click())

        // Then
        onView(withId(R.id.user_name))
            .check(matches(isDisplayed()))
            .check(matches(withText(user.name)))

        coVerify(exactly = 2) { mockRepository.getUser("123") }
    }

    @Test
    fun `fragment handles configuration change during error state`() {
        // Given
        coEvery { mockRepository.getUser(any()) } returns Result.failure(Exception("Error"))

        scenario.onFragment { fragment ->
            fragment.loadUser("123")
        }

        // When
        scenario.recreate()

        // Then
        onView(withId(R.id.error_message))
            .check(matches(isDisplayed()))
    }
}
```

#### Testing Network Error Scenarios with MockWebServer

```kotlin
class NetworkErrorIntegrationTest {

    private lateinit var mockWebServer: MockWebServer
    private lateinit var apiService: ApiService

    @Before
    fun setup() {
        mockWebServer = MockWebServer()
        mockWebServer.start()

        apiService = ApiService.create(mockWebServer.url("/").toString())
    }

    @After
    fun tearDown() {
        mockWebServer.shutdown()
    }

    @Test
    fun `handles server error responses correctly`() = runTest {
        // Given
        mockWebServer.enqueue(
            MockResponse()
                .setResponseCode(500)
                .setBody("Internal Server Error")
        )

        // When
        val result = apiService.getUser("123")

        // Then
        assertThat(result.isFailure).isTrue()
        val exception = result.exceptionOrNull() as? HttpException
        assertThat(exception?.code()).isEqualTo(500)
    }

    @Test
    fun `handles network timeout correctly`() = runTest {
        // Given
        mockWebServer.enqueue(
            MockResponse()
                .setSocketPolicy(SocketPolicy.NO_RESPONSE)
        )

        // When
        val result = apiService.getUser("123")

        // Then
        assertThat(result.isFailure).isTrue()
        assertThat(result.exceptionOrNull()).isInstanceOf(SocketTimeoutException::class.java)
    }

    @Test
    fun `retry logic works with temporary failures`() = runTest {
        // Given - first two requests fail, third succeeds
        mockWebServer.enqueue(MockResponse().setResponseCode(500))
        mockWebServer.enqueue(MockResponse().setResponseCode(500))
        mockWebServer.enqueue(
            MockResponse()
                .setResponseCode(200)
                .setBody("""{"id":"123","name":"John Doe","email":"john@example.com"}""")
        )

        val retryService = RetryService(apiService)

        // When
        val result = retryService.getUserWithRetry("123")

        // Then
        assertThat(result.isSuccess).isTrue()
        assertThat(mockWebServer.requestCount).isEqualTo(3)
    }
}
```

### 3. UI Error State Testing with Espresso

```kotlin
@RunWith(AndroidJUnit4::class)
@LargeTest
class ErrorHandlingUITest {

    @get:Rule
    val activityRule = ActivityScenarioRule(MainActivity::class.java)

    @Before
    fun setup() {
        // Setup test dependencies
        IdlingRegistry.getInstance().register(EspressoIdlingResource.countingIdlingResource)
    }

    @After
    fun tearDown() {
        IdlingRegistry.getInstance().unregister(EspressoIdlingResource.countingIdlingResource)
    }

    @Test
    fun displayNetworkErrorAndAllowRetry() {
        // Given - Setup network failure scenario
        TestNetworkManager.simulateNetworkFailure()

        // When - Navigate to user profile
        onView(withId(R.id.nav_profile))
            .perform(click())

        // Then - Verify error state is displayed
        onView(withId(R.id.error_container))
            .check(matches(isDisplayed()))

        onView(withId(R.id.error_title))
            .check(matches(withText("Connection Problem")))

        onView(withId(R.id.error_message))
            .check(matches(withText(containsString("Check your internet connection"))))

        onView(withId(R.id.retry_button))
            .check(matches(isDisplayed()))
            .check(matches(isEnabled()))

        // When - Fix network and retry
        TestNetworkManager.restoreNetwork()
        onView(withId(R.id.retry_button))
            .perform(click())

        // Then - Verify successful state
        onView(withId(R.id.user_profile_container))
            .check(matches(isDisplayed()))

        onView(withId(R.id.error_container))
            .check(matches(not(isDisplayed())))
    }

    @Test
    fun displayPermissionErrorWithSettingsRedirect() {
        // Given - Simulate location permission denied
        PermissionTestUtils.denyLocationPermission()

        // When - Request location-based feature
        onView(withId(R.id.find_nearby_button))
            .perform(click())

        // Then - Verify permission error dialog
        onView(withText("Location Permission Required"))
            .check(matches(isDisplayed()))

        onView(withText(containsString("location permission")))
            .check(matches(isDisplayed()))

        onView(withText("Open Settings"))
            .check(matches(isDisplayed()))
            .perform(click())

        // Verify settings intent was launched
        intended(hasAction(Settings.ACTION_APPLICATION_DETAILS_SETTINGS))
    }

    @Test
    fun displayOfflineModeWhenNetworkUnavailable() {
        // Given - No network connection
        TestNetworkManager.disableNetwork()

        // When - Try to load data
        onView(withId(R.id.refresh_button))
            .perform(click())

        // Then - Verify offline mode UI
        onView(withId(R.id.offline_indicator))
            .check(matches(isDisplayed()))

        onView(withText("Offline Mode"))
            .check(matches(isDisplayed()))

        onView(withText(containsString("cached data")))
            .check(matches(isDisplayed()))

        // Verify cached data is shown
        onView(withId(R.id.data_container))
            .check(matches(isDisplayed()))
    }
}
```

### 4. Testing Error Handling with Compose

```kotlin
class ComposeErrorHandlingTest {

    @get:Rule
    val composeTestRule = createComposeRule()

    @Test
    fun errorStateDisplaysCorrectContent() {
        // Given
        val errorState = UiState.Error("Network connection failed")

        // When
        composeTestRule.setContent {
            UserScreen(
                uiState = errorState,
                onRetry = { }
            )
        }

        // Then
        composeTestRule
            .onNodeWithText("Network connection failed")
            .assertIsDisplayed()

        composeTestRule
            .onNodeWithText("Retry")
            .assertIsDisplayed()
            .assertHasClickAction()
    }

    @Test
    fun retryButtonTriggersCorrectCallback() {
        // Given
        var retryClicked = false
        val errorState = UiState.Error("Test error")

        // When
        composeTestRule.setContent {
            UserScreen(
                uiState = errorState,
                onRetry = { retryClicked = true }
            )
        }

        composeTestRule
            .onNodeWithText("Retry")
            .performClick()

        // Then
        assertThat(retryClicked).isTrue()
    }

    @Test
    fun loadingStateShowsProgressIndicator() {
        // Given
        val loadingState = UiState.Loading

        // When
        composeTestRule.setContent {
            UserScreen(
                uiState = loadingState,
                onRetry = { }
            )
        }

        // Then
        composeTestRule
            .onNode(hasTestTag("loading_indicator"))
            .assertIsDisplayed()
    }
}
```

### 5. Test Utilities and Helpers

```kotlin
// Test utilities for error scenarios
object ErrorTestUtils {

    fun createNetworkException(message: String = "Network error"): IOException {
        return IOException(message)
    }

    fun createHttpException(code: Int, message: String = "HTTP error"): HttpException {
        val response = Response.error<Any>(
            code,
            message.toResponseBody("text/plain".toMediaTypeOrNull())
        )
        return HttpException(response)
    }

    fun createDatabaseException(message: String = "Database error"): SQLiteException {
        return SQLiteException(message)
    }
}

// Mock response builders
object MockResponseBuilder {

    fun successResponse(body: String): MockResponse {
        return MockResponse()
            .setResponseCode(200)
            .setBody(body)
            .addHeader("Content-Type", "application/json")
    }

    fun errorResponse(code: Int, message: String): MockResponse {
        return MockResponse()
            .setResponseCode(code)
            .setBody("""{"error": "$message"}""")
    }

    fun timeoutResponse(): MockResponse {
        return MockResponse()
            .setSocketPolicy(SocketPolicy.NO_RESPONSE)
    }

    fun slowResponse(delayMs: Long, body: String): MockResponse {
        return MockResponse()
            .setResponseCode(200)
            .setBody(body)
            .setBodyDelayTimeMs(delayMs)
    }
}

// Test doubles for error scenarios
class TestUserRepository : UserRepository {

    var shouldFail = false
    var failureException: Exception = RuntimeException("Test failure")
    var responseDelay = 0L

    override suspend fun getUser(userId: String): Result<User> {
        if (responseDelay > 0) {
            delay(responseDelay)
        }

        return if (shouldFail) {
            Result.failure(failureException)
        } else {
            Result.success(createTestUser(userId))
        }
    }

    fun simulateNetworkError() {
        shouldFail = true
        failureException = IOException("Network connection failed")
    }

    fun simulateServerError() {
        shouldFail = true
        failureException = HttpException(
            Response.error<Any>(500, "Server Error".toResponseBody())
        )
    }

    fun simulateTimeout() {
        responseDelay = 10000
    }

    fun reset() {
        shouldFail = false
        responseDelay = 0
    }
}
```

**Testing Best Practices:**
- Test both success and failure paths
- Use appropriate test doubles (mocks, fakes, stubs)
- Test error recovery mechanisms
- Verify user-facing error messages
- Test configuration changes during error states
- Use parameterized tests for multiple error scenarios
- Test timeout and cancellation behavior

---

## Error Logging and Monitoring

Effective error logging and monitoring are crucial for maintaining mobile application quality and user experience.

### 1. Firebase Crashlytics Integration

#### Basic Setup and Configuration

```kotlin
class CrashlyticsManager {
    private val crashlytics = FirebaseCrashlytics.getInstance()

    fun initialize() {
        // Set custom keys for better error context
        crashlytics.setCustomKey("user_type", getCurrentUserType())
        crashlytics.setCustomKey("app_version", BuildConfig.VERSION_NAME)
        crashlytics.setCustomKey("api_level", Build.VERSION.SDK_INT)
    }

    fun setUserContext(user: User) {
        crashlytics.setUserId(user.id)
        crashlytics.setCustomKey("user_tier", user.subscriptionTier)
        crashlytics.setCustomKey("user_region", user.region)
    }

    fun logHandledException(exception: Throwable, context: String) {
        // Add context breadcrumbs
        crashlytics.log("Context: $context")
        crashlytics.recordException(exception)
    }

    fun addBreadcrumb(message: String) {
        crashlytics.log(message)
    }

    fun logCustomError(
        message: String,
        category: String,
        severity: Priority = Priority.HIGH
    ) {
        crashlytics.setCustomKey("error_category", category)
        crashlytics.setCustomKey("error_severity", severity.name)
        crashlytics.recordException(CustomException(message, category))
    }

    enum class Priority {
        LOW, MEDIUM, HIGH, CRITICAL
    }

    class CustomException(
        message: String,
        val category: String
    ) : Exception(message)
}
```

#### Advanced Error Context Collection

```kotlin
class ErrorContextCollector {

    fun collectErrorContext(exception: Throwable): ErrorContext {
        return ErrorContext(
            timestamp = System.currentTimeMillis(),
            exception = exception,
            userInfo = collectUserInfo(),
            deviceInfo = collectDeviceInfo(),
            appInfo = collectAppInfo(),
            networkInfo = collectNetworkInfo(),
            memoryInfo = collectMemoryInfo()
        )
    }

    private fun collectDeviceInfo(): DeviceInfo {
        return DeviceInfo(
            manufacturer = Build.MANUFACTURER,
            model = Build.MODEL,
            androidVersion = Build.VERSION.RELEASE,
            apiLevel = Build.VERSION.SDK_INT,
            screenDensity = Resources.getSystem().displayMetrics.density,
            availableStorage = getAvailableStorage(),
            batteryLevel = getBatteryLevel()
        )
    }

    private fun collectAppInfo(): AppInfo {
        return AppInfo(
            versionName = BuildConfig.VERSION_NAME,
            versionCode = BuildConfig.VERSION_CODE,
            buildType = BuildConfig.BUILD_TYPE,
            flavor = BuildConfig.FLAVOR,
            installSource = getInstallSource()
        )
    }

    private fun collectNetworkInfo(): NetworkInfo {
        val connectivityManager = context.getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager
        val activeNetwork = connectivityManager.activeNetworkInfo

        return NetworkInfo(
            isConnected = activeNetwork?.isConnected ?: false,
            networkType = activeNetwork?.type?.toString() ?: "unknown",
            isWifi = activeNetwork?.type == ConnectivityManager.TYPE_WIFI,
            isMobile = activeNetwork?.type == ConnectivityManager.TYPE_MOBILE
        )
    }

    private fun collectMemoryInfo(): MemoryInfo {
        val runtime = Runtime.getRuntime()
        val activityManager = context.getSystemService(Context.ACTIVITY_SERVICE) as ActivityManager
        val memoryInfo = ActivityManager.MemoryInfo()
        activityManager.getMemoryInfo(memoryInfo)

        return MemoryInfo(
            usedMemory = runtime.totalMemory() - runtime.freeMemory(),
            totalMemory = runtime.totalMemory(),
            maxMemory = runtime.maxMemory(),
            availableSystemMemory = memoryInfo.availMem,
            lowMemory = memoryInfo.lowMemory
        )
    }
}
```

### 2. Structured Logging Implementation

```kotlin
class StructuredLogger {

    companion object {
        private const val TAG_PREFIX = "MyApp"

        // Log levels
        const val LEVEL_VERBOSE = Log.VERBOSE
        const val LEVEL_DEBUG = Log.DEBUG
        const val LEVEL_INFO = Log.INFO
        const val LEVEL_WARN = Log.WARN
        const val LEVEL_ERROR = Log.ERROR
    }

    fun logError(
        tag: String,
        message: String,
        throwable: Throwable? = null,
        additionalData: Map<String, Any> = emptyMap()
    ) {
        val logEntry = createLogEntry(
            level = LEVEL_ERROR,
            tag = tag,
            message = message,
            throwable = throwable,
            additionalData = additionalData
        )

        Log.e("$TAG_PREFIX:$tag", logEntry.toJson())

        // Send to crash reporting
        CrashlyticsManager.instance.logHandledException(
            throwable ?: RuntimeException(message),
            tag
        )

        // Send to analytics if needed
        if (shouldReportToAnalytics(throwable)) {
            sendToAnalytics(logEntry)
        }
    }

    fun logWarning(
        tag: String,
        message: String,
        additionalData: Map<String, Any> = emptyMap()
    ) {
        val logEntry = createLogEntry(
            level = LEVEL_WARN,
            tag = tag,
            message = message,
            additionalData = additionalData
        )

        Log.w("$TAG_PREFIX:$tag", logEntry.toJson())
    }

    fun logInfo(
        tag: String,
        message: String,
        additionalData: Map<String, Any> = emptyMap()
    ) {
        val logEntry = createLogEntry(
            level = LEVEL_INFO,
            tag = tag,
            message = message,
            additionalData = additionalData
        )

        Log.i("$TAG_PREFIX:$tag", logEntry.toJson())
    }

    private fun createLogEntry(
        level: Int,
        tag: String,
        message: String,
        throwable: Throwable? = null,
        additionalData: Map<String, Any> = emptyMap()
    ): LogEntry {
        return LogEntry(
            timestamp = System.currentTimeMillis(),
            level = getLevelString(level),
            tag = tag,
            message = message,
            exception = throwable?.let {
                ExceptionInfo(
                    type = it.javaClass.simpleName,
                    message = it.message,
                    stackTrace = it.stackTraceToString()
                )
            },
            additionalData = additionalData,
            threadName = Thread.currentThread().name,
            sessionId = SessionManager.getCurrentSessionId()
        )
    }

    data class LogEntry(
        val timestamp: Long,
        val level: String,
        val tag: String,
        val message: String,
        val exception: ExceptionInfo?,
        val additionalData: Map<String, Any>,
        val threadName: String,
        val sessionId: String
    ) {
        fun toJson(): String {
            return Gson().toJson(this)
        }
    }

    data class ExceptionInfo(
        val type: String,
        val message: String?,
        val stackTrace: String
    )
}
```

### 3. Repository Error Logging Pattern

```kotlin
class LoggingUserRepository(
    private val userApi: UserApi,
    private val logger: StructuredLogger
) : UserRepository {

    override suspend fun getUser(userId: String): Result<User> {
        return try {
            logger.logInfo(
                tag = "UserRepository",
                message = "Fetching user data",
                additionalData = mapOf("userId" to userId)
            )

            val user = userApi.getUser(userId)

            logger.logInfo(
                tag = "UserRepository",
                message = "Successfully fetched user data",
                additionalData = mapOf(
                    "userId" to userId,
                    "userTier" to user.tier,
                    "responseTime" to measureTimeMillis { userApi.getUser(userId) }
                )
            )

            Result.success(user)

        } catch (e: Exception) {
            logger.logError(
                tag = "UserRepository",
                message = "Failed to fetch user data",
                throwable = e,
                additionalData = mapOf(
                    "userId" to userId,
                    "errorType" to e.javaClass.simpleName,
                    "endpoint" to "/users/$userId"
                )
            )

            Result.failure(mapException(e))
        }
    }

    private fun mapException(exception: Exception): Exception {
        return when (exception) {
            is IOException -> {
                logger.logError(
                    tag = "NetworkError",
                    message = "Network connectivity issue",
                    throwable = exception,
                    additionalData = mapOf(
                        "networkType" to getNetworkType(),
                        "isConnected" to isNetworkConnected()
                    )
                )
                NetworkException("Network error occurred", exception)
            }

            is HttpException -> {
                logger.logError(
                    tag = "ApiError",
                    message = "HTTP error response",
                    throwable = exception,
                    additionalData = mapOf(
                        "statusCode" to exception.code(),
                        "responseMessage" to exception.message(),
                        "endpoint" to exception.response()?.raw()?.request?.url?.toString()
                    )
                )

                when (exception.code()) {
                    404 -> UserNotFoundException("User not found")
                    401 -> AuthenticationException("Authentication required")
                    else -> ApiException("API error: ${exception.code()}")
                }
            }

            else -> exception
        }
    }
}
```

### 4. Performance Monitoring Integration

```kotlin
class PerformanceMonitor {

    private val performanceMonitoring = FirebasePerformance.getInstance()

    fun traceNetworkRequest(
        name: String,
        httpMethod: String,
        url: String
    ): Trace {
        return performanceMonitoring.newTrace("network_$name").apply {
            putAttribute("http_method", httpMethod)
            putAttribute("url", url)
            start()
        }
    }

    fun traceCustomOperation(name: String): Trace {
        return performanceMonitoring.newTrace(name).apply {
            start()
        }
    }

    suspend fun <T> measureSuspendFunction(
        operationName: String,
        operation: suspend () -> T
    ): T {
        val trace = traceCustomOperation(operationName)
        return try {
            val result = operation()
            trace.putAttribute("success", "true")
            result
        } catch (e: Exception) {
            trace.putAttribute("success", "false")
            trace.putAttribute("error_type", e.javaClass.simpleName)
            throw e
        } finally {
            trace.stop()
        }
    }

    fun recordScreenTrace(screenName: String, loadTimeMs: Long) {
        val trace = performanceMonitoring.newTrace("screen_$screenName")
        trace.putMetric("load_time_ms", loadTimeMs)
        trace.start()
        trace.stop()
    }
}

// Usage in repository
class MonitoredUserRepository : UserRepository {

    private val performanceMonitor = PerformanceMonitor()

    override suspend fun getUser(userId: String): Result<User> {
        return performanceMonitor.measureSuspendFunction("get_user") {
            val trace = performanceMonitor.traceNetworkRequest(
                name = "get_user",
                httpMethod = "GET",
                url = "/users/$userId"
            )

            try {
                val user = userApi.getUser(userId)
                trace.putAttribute("user_tier", user.tier)
                Result.success(user)
            } catch (e: Exception) {
                trace.putAttribute("error", e.javaClass.simpleName)
                throw e
            } finally {
                trace.stop()
            }
        }
    }
}
```

### 5. Custom Analytics for Error Patterns

```kotlin
class ErrorAnalytics {

    private val firebaseAnalytics = FirebaseAnalytics.getInstance(context)

    fun trackError(
        errorType: String,
        errorCode: String? = null,
        fatal: Boolean = false,
        additionalParams: Map<String, String> = emptyMap()
    ) {
        val bundle = Bundle().apply {
            putString("error_type", errorType)
            errorCode?.let { putString("error_code", it) }
            putBoolean("fatal", fatal)

            additionalParams.forEach { (key, value) ->
                putString(key, value)
            }
        }

        firebaseAnalytics.logEvent("app_error", bundle)
    }

    fun trackErrorRecovery(
        errorType: String,
        recoveryMethod: String,
        successful: Boolean,
        attemptNumber: Int = 1
    ) {
        val bundle = Bundle().apply {
            putString("error_type", errorType)
            putString("recovery_method", recoveryMethod)
            putBoolean("recovery_successful", successful)
            putInt("attempt_number", attemptNumber)
        }

        firebaseAnalytics.logEvent("error_recovery", bundle)
    }

    fun trackUserExperience(
        feature: String,
        errorEncountered: Boolean,
        completionTime: Long,
        userSatisfaction: Int? = null
    ) {
        val bundle = Bundle().apply {
            putString("feature", feature)
            putBoolean("error_encountered", errorEncountered)
            putLong("completion_time_ms", completionTime)
            userSatisfaction?.let { putInt("satisfaction_rating", it) }
        }

        firebaseAnalytics.logEvent("user_experience", bundle)
    }
}

// Integration with error handling
class AnalyticsAwareErrorHandler {

    private val errorAnalytics = ErrorAnalytics()

    fun handleApiError(exception: HttpException, endpoint: String) {
        errorAnalytics.trackError(
            errorType = "api_error",
            errorCode = exception.code().toString(),
            fatal = false,
            additionalParams = mapOf(
                "endpoint" to endpoint,
                "response_message" to (exception.message() ?: "")
            )
        )

        // Attempt recovery
        val recoverySuccessful = attemptErrorRecovery(exception)

        errorAnalytics.trackErrorRecovery(
            errorType = "api_error",
            recoveryMethod = "cache_fallback",
            successful = recoverySuccessful
        )
    }

    private fun attemptErrorRecovery(exception: HttpException): Boolean {
        return try {
            // Implement recovery logic
            true
        } catch (e: Exception) {
            false
        }
    }
}
```

### 6. Centralized Error Reporting

```kotlin
class CentralizedErrorReporter {

    private val crashlytics = CrashlyticsManager.instance
    private val logger = StructuredLogger()
    private val analytics = ErrorAnalytics()
    private val performanceMonitor = PerformanceMonitor()

    fun reportError(
        error: AppError,
        context: String,
        severity: ErrorSeverity = ErrorSeverity.MEDIUM
    ) {
        // Structured logging
        logger.logError(
            tag = error.category,
            message = error.message,
            throwable = error.throwable,
            additionalData = error.metadata
        )

        // Crash reporting
        crashlytics.setCustomKey("error_category", error.category)
        crashlytics.setCustomKey("error_severity", severity.name)
        crashlytics.setCustomKey("error_context", context)
        crashlytics.recordException(error.throwable ?: RuntimeException(error.message))

        // Analytics
        analytics.trackError(
            errorType = error.category,
            errorCode = error.code,
            fatal = severity == ErrorSeverity.CRITICAL,
            additionalParams = error.metadata.mapValues { it.value.toString() }
        )

        // Performance impact tracking
        if (error.performanceImpact > 0) {
            val trace = performanceMonitor.traceCustomOperation("error_recovery")
            trace.putMetric("performance_impact_ms", error.performanceImpact)
            trace.stop()
        }
    }

    data class AppError(
        val category: String,
        val message: String,
        val code: String? = null,
        val throwable: Throwable? = null,
        val metadata: Map<String, Any> = emptyMap(),
        val performanceImpact: Long = 0
    )

    enum class ErrorSeverity {
        LOW, MEDIUM, HIGH, CRITICAL
    }
}
```

**Error Logging Best Practices:**
- Include contextual information (user ID, session ID, app state)
- Use structured logging formats for better analysis
- Set up proper log retention and privacy policies
- Monitor error trends and patterns over time
- Implement performance monitoring alongside error tracking
- Create alerts for critical error patterns
- Regularly review and analyze error reports for improvements

---

## Conclusion

This comprehensive guide covers the essential aspects of error handling in Kotlin for mobile development. From basic exception handling to advanced reactive patterns, the techniques outlined here provide a robust foundation for building reliable Android applications.

### Key Takeaways

1. **Layer Your Error Handling**: Combine multiple approaches for comprehensive coverage
2. **Make Errors Explicit**: Use sealed classes and Result types to make errors part of your type system
3. **Handle Mobile-Specific Scenarios**: Account for network connectivity, permissions, and lifecycle issues
4. **Test Error Scenarios**: Ensure your error handling works correctly under various failure conditions
5. **Monitor and Learn**: Use proper logging and monitoring to continuously improve your error handling

### Next Steps

- Implement structured error handling in your current projects
- Set up comprehensive error monitoring and logging
- Create reusable error handling utilities for your team
- Regularly review and update error handling patterns based on real-world usage
- Stay updated with the latest Kotlin and Android error handling best practices

### Additional Resources

- [Kotlin Coroutines Exception Handling Documentation](https://kotlinlang.org/docs/exception-handling.html)
- [Android Architecture Components Error Handling](https://developer.android.com/topic/libraries/architecture/guide)
- [Arrow Functional Programming Library](https://arrow-kt.io/)
- [Firebase Crashlytics Documentation](https://firebase.google.com/docs/crashlytics)
- [Android Testing Documentation](https://developer.android.com/training/testing)

---

*This document serves as a comprehensive reference for error handling patterns in Kotlin mobile development. Regular updates ensure it remains current with the latest best practices and platform changes.*