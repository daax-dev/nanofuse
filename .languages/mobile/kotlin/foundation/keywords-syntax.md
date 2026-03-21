# Kotlin Keywords and Syntax Reference

*Complete reference for Kotlin keywords, operators, and syntax with mobile development focus*

## Table of Contents

1. [Hard Keywords](#hard-keywords)
2. [Soft Keywords](#soft-keywords)
3. [Modifier Keywords](#modifier-keywords)
4. [Special Identifiers](#special-identifiers)
5. [Operators and Special Symbols](#operators-and-special-symbols)
6. [Mobile Development Syntax Patterns](#mobile-development-syntax-patterns)
7. [Literals and Constants](#literals-and-constants)
8. [Grammar Rules](#grammar-rules)

---

## Hard Keywords

*These tokens are always interpreted as keywords and cannot be used as identifiers*

### Type System and Casting

| Keyword | Description | Mobile Example |
|---------|-------------|----------------|
| `as` | Type casting, import aliasing | `val bitmap = drawable as BitmapDrawable` |
| `as?` | Safe type casting (returns null if fails) | `val user = data as? User` |
| `is` | Type checking | `if (result is NetworkError)` |
| `!is` | Negated type checking | `if (response !is Success)` |

```kotlin
// Mobile example: Safe casting in network responses
when (val response = apiCall()) {
    is Success -> handleSuccess(response.data)
    is Error -> handleError(response.message)
    else -> handleUnknown()
}

// Safe casting for Android views
val editText = findViewById(R.id.input) as? EditText
    ?: throw IllegalStateException("View is not EditText")
```

### Control Flow

| Keyword | Description | Mobile Example |
|---------|-------------|----------------|
| `if` | Conditional expression | `if (isNetworkAvailable()) fetchData()` |
| `else` | Alternative branch | `else showCachedData()` |
| `when` | Pattern matching (replaces switch) | `when (connectionState) { ... }` |
| `for` | Loop iteration | `for (item in recyclerItems)` |
| `while` | While loop | `while (isLoading)` |
| `do` | Do-while loop | `do { retry() } while (!success)` |
| `break` | Exit loop | `break@retryLoop` |
| `continue` | Skip to next iteration | `continue` |

```kotlin
// Mobile example: Connection state handling
when (networkState) {
    NetworkState.CONNECTED -> {
        syncData()
        showOnlineUI()
    }
    NetworkState.CONNECTING -> showLoadingIndicator()
    NetworkState.DISCONNECTED -> {
        showOfflineMode()
        enableCachedData()
    }
}

// Android lifecycle-aware loops
for (fragment in supportFragmentManager.fragments) {
    if (fragment.isDetached) continue
    fragment.onResume()
}
```

### Declarations

| Keyword | Description | Mobile Example |
|---------|-------------|----------------|
| `class` | Class declaration | `class MainActivity : AppCompatActivity()` |
| `interface` | Interface declaration | `interface ApiService` |
| `object` | Singleton object | `object DatabaseManager` |
| `fun` | Function declaration | `fun fetchUserProfile()` |
| `val` | Read-only property | `val userId = intent.getStringExtra("id")` |
| `var` | Mutable property | `var isLoading = false` |

```kotlin
// Mobile example: Android component declarations
class UserProfileFragment : Fragment() {
    private val viewModel: UserProfileViewModel by viewModels()
    private var binding: FragmentUserProfileBinding? = null

    fun loadUserData() {
        viewModel.fetchProfile(userId)
    }
}

// Singleton for mobile app configuration
object AppConfig {
    const val API_BASE_URL = "https://api.example.com"
    val isDebugMode = BuildConfig.DEBUG
}
```

### Memory and References

| Keyword | Description | Mobile Example |
|---------|-------------|----------------|
| `null` | Null reference | `if (user != null)` |
| `this` | Current instance reference | `this.findViewById(R.id.button)` |
| `super` | Superclass reference | `super.onCreate(savedInstanceState)` |
| `return` | Return from function | `return repository.getUser(id)` |
| `throw` | Throw exception | `throw NetworkException("Connection failed")` |
| `try` | Exception handling | `try { parseResponse() }` |

```kotlin
// Mobile example: Android lifecycle with proper super calls
class MainActivity : AppCompatActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState) // Required super call
        setContentView(R.layout.activity_main)
    }

    private fun handleApiError(): User? {
        return try {
            apiService.getCurrentUser()
        } catch (e: NetworkException) {
            showErrorDialog(e.message)
            null
        }
    }
}
```

### Collections and Ranges

| Keyword | Description | Mobile Example |
|---------|-------------|----------------|
| `in` | Membership check, loop iteration | `if (position in 0..itemCount)` |
| `!in` | Negated membership check | `if (userId !in blockedUsers)` |

```kotlin
// Mobile example: Range checks for RecyclerView
class RecyclerAdapter : RecyclerView.Adapter<ViewHolder>() {
    override fun onBindViewHolder(holder: ViewHolder, position: Int) {
        if (position !in 0 until items.size) return
        holder.bind(items[position])
    }
}

// Check if network type is supported
val supportedNetworkTypes = listOf(
    ConnectivityManager.TYPE_WIFI,
    ConnectivityManager.TYPE_MOBILE
)
if (networkType in supportedNetworkTypes) {
    // Proceed with network operation
}
```

### Type System

| Keyword | Description | Mobile Example |
|---------|-------------|----------------|
| `typealias` | Type alias | `typealias UserId = String` |
| `typeof` | Reserved for future use | N/A |

### Literals

| Keyword | Description | Value |
|---------|-------------|-------|
| `true` | Boolean true | `true` |
| `false` | Boolean false | `false` |

### Package Declaration

| Keyword | Description | Mobile Example |
|---------|-------------|----------------|
| `package` | Package declaration | `package com.example.myapp.ui` |

---

## Soft Keywords

*These have special meaning in specific contexts but can be used as identifiers elsewhere*

### Property and Delegation

| Keyword | Context | Description | Mobile Example |
|---------|---------|-------------|----------------|
| `by` | Delegation | Delegates implementation | `val user by lazy { getUser() }` |
| `get` | Property getter | Custom getter | `get() = preferences.getString("key", "")` |
| `set` | Property setter | Custom setter | `set(value) { preferences.edit().putString("key", value).apply() }` |
| `field` | Property backing field | Access backing field | `set(value) { field = value; notifyChange() }` |

```kotlin
// Mobile example: Lazy initialization for expensive operations
class UserRepository {
    private val database by lazy {
        Room.databaseBuilder(context, AppDatabase::class.java, "app-db").build()
    }

    var currentUser: User? = null
        set(value) {
            field = value
            preferences.edit()
                .putString("current_user_id", value?.id)
                .apply()
        }
        get() = field ?: loadUserFromPreferences()
}

// Property delegation for Android preferences
class SettingsManager(private val preferences: SharedPreferences) {
    var isNotificationsEnabled by SharedPreferencesDelegate(
        preferences, "notifications_enabled", true
    )
}
```

### Exception Handling

| Keyword | Context | Description | Mobile Example |
|---------|---------|-------------|----------------|
| `catch` | Exception handling | Catch specific exceptions | `catch (e: NetworkException)` |
| `finally` | Exception handling | Always executed block | `finally { hideLoadingIndicator() }` |

```kotlin
// Mobile example: Network operation with proper cleanup
suspend fun fetchUserData(userId: String): Result<User> {
    showLoadingIndicator()
    return try {
        val user = apiService.getUser(userId)
        Result.success(user)
    } catch (e: NetworkException) {
        logError("Network error", e)
        Result.failure(e)
    } catch (e: SecurityException) {
        logError("Security error", e)
        Result.failure(e)
    } finally {
        hideLoadingIndicator()
    }
}
```

### Class Construction

| Keyword | Context | Description | Mobile Example |
|---------|---------|-------------|----------------|
| `constructor` | Class constructor | Primary/secondary constructor | `constructor(context: Context)` |
| `init` | Class initialization | Initialization block | `init { setupViews() }` |

```kotlin
// Mobile example: Custom View with proper initialization
class CustomProgressBar @JvmOverloads constructor(
    context: Context,
    attrs: AttributeSet? = null,
    defStyleAttr: Int = 0
) : View(context, attrs, defStyleAttr) {

    private var progress: Float = 0f

    init {
        setupPaint()
        setupAnimations()

        attrs?.let { attributeSet ->
            val typedArray = context.obtainStyledAttributes(
                attributeSet, R.styleable.CustomProgressBar
            )
            try {
                progress = typedArray.getFloat(R.styleable.CustomProgressBar_progress, 0f)
            } finally {
                typedArray.recycle()
            }
        }
    }

    // Secondary constructor for programmatic creation
    constructor(context: Context, initialProgress: Float) : this(context) {
        this.progress = initialProgress
    }
}
```

### Import and File Organization

| Keyword | Context | Description | Mobile Example |
|---------|---------|-------------|----------------|
| `import` | File imports | Import declarations | `import androidx.lifecycle.ViewModel` |
| `where` | Generic constraints | Type parameter constraints | `fun <T> save(item: T) where T : Serializable` |

### Loop Control

| Keyword | Context | Description | Mobile Example |
|---------|---------|-------------|----------------|
| `while` | Loop control | While loop condition | `while (isConnected && hasMoreData)` |

### Annotation Targets

| Keyword | Context | Description | Mobile Example |
|---------|---------|-------------|----------------|
| `delegate` | Annotation target | Target delegate | `@delegate:Transient` |
| `file` | Annotation target | Target entire file | `@file:JvmName("Utils")` |
| `param` | Annotation target | Target parameter | `@param:NotNull` |
| `property` | Annotation target | Target property | `@property:SerializedName("user_id")` |
| `receiver` | Annotation target | Target receiver | `@receiver:NotNull` |
| `setparam` | Annotation target | Target setter parameter | `@setparam:Inject` |

### Special Types

| Keyword | Context | Description | Mobile Example |
|---------|---------|-------------|----------------|
| `dynamic` | Kotlin/JS | Dynamic type | Not applicable for mobile |
| `value` | Inline classes | Value class declaration | `@JvmInline value class UserId(val id: String)` |

---

## Modifier Keywords

*Used in declaration lists to modify behavior or visibility*

### Visibility Modifiers

| Modifier | Scope | Description | Mobile Example |
|----------|-------|-------------|----------------|
| `public` | Everywhere | Public visibility (default) | `public fun loginUser()` |
| `private` | Same class/file | Private to current scope | `private fun validateInput()` |
| `protected` | Subclasses | Visible in subclasses | `protected fun onUserChanged()` |
| `internal` | Same module | Module-internal visibility | `internal class DatabaseManager` |

### Class Modifiers

| Modifier | Description | Mobile Example |
|----------|-------------|----------------|
| `abstract` | Cannot be instantiated | `abstract class BaseFragment` |
| `open` | Can be subclassed | `open class BaseRepository` |
| `final` | Cannot be subclassed (default) | `final class UserManager` |
| `sealed` | Restricted subclassing | `sealed class NetworkResult` |
| `data` | Data class with generated methods | `data class User(val id: String, val name: String)` |
| `enum` | Enumeration | `enum class Theme { LIGHT, DARK, AUTO }` |
| `annotation` | Annotation class | `annotation class DatabaseEntity` |
| `inner` | Access to outer class instance | `inner class ViewHolder` |
| `companion` | Companion object | `companion object { const val TAG = "MainActivity" }` |

```kotlin
// Mobile example: Sealed class for network results
sealed class ApiResult<out T> {
    data class Success<T>(val data: T) : ApiResult<T>()
    data class Error(val exception: Throwable) : ApiResult<Nothing>()
    object Loading : ApiResult<Nothing>()
}

// Data class for API responses
data class UserProfile(
    @SerializedName("user_id") val userId: String,
    @SerializedName("display_name") val displayName: String,
    @SerializedName("avatar_url") val avatarUrl: String?
)

// Abstract base for all fragments
abstract class BaseFragment : Fragment() {
    protected abstract fun setupViews()
    protected abstract fun observeData()

    final override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        setupViews()
        observeData()
    }
}
```

### Function Modifiers

| Modifier | Description | Mobile Example |
|----------|-------------|----------------|
| `override` | Override superclass member | `override fun onCreate()` |
| `inline` | Inline function | `inline fun <T> measureTime(block: () -> T)` |
| `suspend` | Suspending function (coroutine) | `suspend fun fetchData()` |
| `tailrec` | Tail-recursive function | `tailrec fun factorial(n: Int, acc: Int = 1)` |
| `operator` | Operator overloading | `operator fun plus(other: Vector)` |
| `infix` | Infix function call | `infix fun String.shouldBe(expected: String)` |
| `external` | External implementation | `external fun nativeCalculation(): Int` |

```kotlin
// Mobile example: Suspending functions for async operations
class UserRepository {
    suspend fun fetchUserProfile(userId: String): User {
        return withContext(Dispatchers.IO) {
            apiService.getUser(userId)
        }
    }

    suspend fun updateUserProfile(user: User): Result<User> {
        return try {
            val updatedUser = apiService.updateUser(user)
            database.userDao().update(updatedUser)
            Result.success(updatedUser)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}

// Inline function for measuring performance
inline fun <T> measureExecutionTime(
    tag: String,
    block: () -> T
): T {
    val startTime = System.currentTimeMillis()
    val result = block()
    val endTime = System.currentTimeMillis()
    Log.d(tag, "Execution time: ${endTime - startTime}ms")
    return result
}

// Operator overloading for custom types
data class Coordinate(val x: Float, val y: Float) {
    operator fun plus(other: Coordinate) = Coordinate(x + other.x, y + other.y)
    operator fun minus(other: Coordinate) = Coordinate(x - other.x, y - other.y)
}
```

### Property Modifiers

| Modifier | Description | Mobile Example |
|----------|-------------|----------------|
| `const` | Compile-time constant | `const val API_TIMEOUT = 30_000L` |
| `lateinit` | Late initialization | `lateinit var binding: ActivityMainBinding` |

```kotlin
// Mobile example: Constants and late initialization
class MainActivity : AppCompatActivity() {
    companion object {
        const val EXTRA_USER_ID = "extra_user_id"
        const val REQUEST_CODE_PERMISSION = 1001
    }

    private lateinit var binding: ActivityMainBinding
    private lateinit var userRepository: UserRepository

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        binding = ActivityMainBinding.inflate(layoutInflater)
        userRepository = (application as MyApplication).userRepository
        setContentView(binding.root)
    }
}
```

### Parameter Modifiers

| Modifier | Description | Mobile Example |
|----------|-------------|----------------|
| `vararg` | Variable arguments | `fun log(tag: String, vararg messages: String)` |
| `noinline` | Don't inline lambda parameter | `fun process(noinline callback: () -> Unit)` |
| `crossinline` | Forbid non-local returns | `fun asyncProcess(crossinline callback: () -> Unit)` |

### Generic Type Modifiers

| Modifier | Description | Mobile Example |
|----------|-------------|----------------|
| `in` | Contravariant type parameter | `interface Comparator<in T>` |
| `out` | Covariant type parameter | `interface Producer<out T>` |
| `reified` | Reified type parameter | `inline fun <reified T> isInstance(obj: Any)` |

### Multiplatform Modifiers

| Modifier | Description | Mobile Example |
|----------|-------------|----------------|
| `expect` | Expected declaration | `expect fun getPlatformName(): String` |
| `actual` | Actual implementation | `actual fun getPlatformName() = "Android"` |

---

## Special Identifiers

*Compiler-defined identifiers with special meaning in specific contexts*

| Identifier | Context | Description | Mobile Example |
|------------|---------|-------------|----------------|
| `field` | Property accessor | Access backing field | `set(value) { field = value; onChange() }` |
| `it` | Lambda parameter | Implicit lambda parameter | `users.filter { it.isActive }` |

```kotlin
// Mobile example: Custom property with backing field
class UserProfileView : LinearLayout {
    var user: User? = null
        set(value) {
            field = value
            updateUI()
            notifyUserChanged()
        }

    private fun updateUI() {
        user?.let { user ->
            nameTextView.text = user.name
            emailTextView.text = user.email
            // Load avatar image
            Glide.with(context)
                .load(user.avatarUrl)
                .into(avatarImageView)
        }
    }
}

// Lambda with implicit 'it' parameter
val activeUsers = allUsers.filter { it.isActive && it.lastLogin != null }
val userNames = activeUsers.map { it.name }
```

---

## Operators and Special Symbols

### Mathematical Operators

| Symbol | Function | Mobile Example |
|--------|----------|----------------|
| `+` | Addition, String concatenation | `"User: " + user.name` |
| `-` | Subtraction | `endTime - startTime` |
| `*` | Multiplication, Vararg | `fun log(vararg *messages: String)` |
| `/` | Division | `progress / totalItems` |
| `%` | Modulo | `position % 2 == 0` |

### Assignment Operators

| Symbol | Function | Mobile Example |
|--------|----------|----------------|
| `=` | Assignment | `isLoading = true` |
| `+=` | Add and assign | `totalCount += newItems.size` |
| `-=` | Subtract and assign | `remainingAttempts -= 1` |
| `*=` | Multiply and assign | `scale *= zoomFactor` |
| `/=` | Divide and assign | `progress /= totalSteps` |
| `%=` | Modulo and assign | `index %= arraySize` |

### Increment/Decrement

| Symbol | Function | Mobile Example |
|--------|----------|----------------|
| `++` | Increment | `++retryCount` |
| `--` | Decrement | `--remainingAttempts` |

### Comparison Operators

| Symbol | Function | Mobile Example |
|--------|----------|----------------|
| `==` | Equality | `if (user.id == currentUserId)` |
| `!=` | Inequality | `if (response != null)` |
| `===` | Referential equality | `if (instance1 === instance2)` |
| `!==` | Referential inequality | `if (oldUser !== newUser)` |
| `<` | Less than | `if (progress < 100)` |
| `>` | Greater than | `if (itemCount > maxItems)` |
| `<=` | Less than or equal | `if (attempts <= maxRetries)` |
| `>=` | Greater than or equal | `if (version >= minRequiredVersion)` |

### Logical Operators

| Symbol | Function | Mobile Example |
|--------|----------|----------------|
| `&&` | Logical AND | `if (isOnline && hasPermission)` |
| `\|\|` | Logical OR | `if (isWifiConnected \|\| isMobileConnected)` |
| `!` | Logical NOT | `if (!isLoading)` |

### Null Safety Operators

| Symbol | Function | Description | Mobile Example |
|--------|----------|-------------|----------------|
| `?` | Nullable type marker | Makes type nullable | `var user: User?` |
| `?.` | Safe call | Call only if not null | `user?.name` |
| `?:` | Elvis operator | Default value if null | `user?.name ?: "Unknown"` |
| `!!` | Not-null assertion | Assert non-null (throws if null) | `user!!.id` |

```kotlin
// Mobile example: Comprehensive null safety
class UserProfileActivity : AppCompatActivity() {
    private var currentUser: User? = null

    private fun displayUserInfo() {
        // Safe call with elvis operator
        val displayName = currentUser?.name ?: "Guest User"
        val email = currentUser?.email ?: "No email provided"

        // Safe call chains
        currentUser?.profile?.avatar?.let { avatarUrl ->
            Glide.with(this)
                .load(avatarUrl)
                .into(binding.avatarImageView)
        }

        // Let function for null checks
        currentUser?.let { user ->
            binding.nameTextView.text = user.name
            binding.emailTextView.text = user.email
            binding.joinDateTextView.text = formatDate(user.joinDate)
        } ?: run {
            // Handle null case
            showLoginPrompt()
        }
    }
}
```

### Range Operators

| Symbol | Function | Mobile Example |
|--------|----------|----------------|
| `..` | Inclusive range | `for (i in 0..10)` |
| `..<` | Exclusive end range | `for (i in 0..<items.size)` |

### Member Access

| Symbol | Function | Mobile Example |
|--------|----------|----------------|
| `.` | Member access | `user.name` |
| `::` | Callable reference | `::handleClick` |

```kotlin
// Mobile example: Method references for event handling
class MainActivity : AppCompatActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        // Method reference for click listener
        binding.loginButton.setOnClickListener(::onLoginClicked)

        // Property reference for observing
        viewModel.userState.observe(this, ::handleUserState)
    }

    private fun onLoginClicked(view: View) {
        // Handle login
    }

    private fun handleUserState(userState: UserState) {
        // Handle state change
    }
}
```

### Array/Collection Access

| Symbol | Function | Mobile Example |
|--------|----------|----------------|
| `[`, `]` | Indexed access | `items[position]` |

### String Templates

| Symbol | Function | Mobile Example |
|--------|----------|----------------|
| `$` | Variable interpolation | `"Hello $name"` |
| `${}` | Expression interpolation | `"Progress: ${current * 100 / total}%"` |

### Annotations and Labels

| Symbol | Function | Mobile Example |
|--------|----------|----------------|
| `@` | Annotations, Labels | `@Override`, `break@loop` |

### Other Symbols

| Symbol | Function | Mobile Example |
|--------|----------|----------------|
| `:` | Type annotation separator | `val name: String` |
| `;` | Statement separator | `doSomething(); doSomethingElse()` |
| `->` | Lambda body separator | `{ item -> item.name }` |
| `_` | Unused parameter placeholder | `items.forEach { _, value -> process(value) }` |

---

## Mobile Development Syntax Patterns

### Android Activity Lifecycle

```kotlin
class MainActivity : AppCompatActivity() {
    private lateinit var binding: ActivityMainBinding
    private val viewModel: MainViewModel by viewModels()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        binding = DataBindingUtil.setContentView(this, R.layout.activity_main)
        binding.lifecycleOwner = this
        binding.viewModel = viewModel

        setupObservers()
        handleIntent(intent)
    }

    override fun onNewIntent(intent: Intent?) {
        super.onNewIntent(intent)
        intent?.let(::handleIntent)
    }

    private fun setupObservers() {
        viewModel.uiState.observe(this) { state ->
            when (state) {
                is UiState.Loading -> showLoading(true)
                is UiState.Success -> {
                    showLoading(false)
                    displayData(state.data)
                }
                is UiState.Error -> {
                    showLoading(false)
                    showError(state.message)
                }
            }
        }
    }
}
```

### Kotlin Coroutines in Android

```kotlin
class UserRepository(
    private val apiService: ApiService,
    private val userDao: UserDao,
    private val ioDispatcher: CoroutineDispatcher = Dispatchers.IO
) {
    suspend fun refreshUser(userId: String): Result<User> = withContext(ioDispatcher) {
        try {
            val user = apiService.getUser(userId)
            userDao.insertUser(user)
            Result.success(user)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    fun getUserFlow(userId: String): Flow<User?> = userDao.getUserFlow(userId)
        .flowOn(ioDispatcher)
        .catch { e ->
            emit(null)
            Log.e("UserRepository", "Error loading user", e)
        }
}
```

### Extension Functions for Android

```kotlin
// Context extensions
fun Context.showToast(message: String, duration: Int = Toast.LENGTH_SHORT) {
    Toast.makeText(this, message, duration).show()
}

fun Context.hideKeyboard(view: View) {
    val inputMethodManager = getSystemService(Context.INPUT_METHOD_SERVICE) as InputMethodManager
    inputMethodManager.hideSoftInputFromWindow(view.windowToken, 0)
}

// View extensions
fun View.visible() {
    visibility = View.VISIBLE
}

fun View.gone() {
    visibility = View.GONE
}

fun View.invisible() {
    visibility = View.INVISIBLE
}

// Fragment extensions
inline fun <reified T : Fragment> FragmentManager.findFragmentByTag(): T? {
    return findFragmentByTag(T::class.java.simpleName) as? T
}

// Usage in Activity
class MainActivity : AppCompatActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        showToast("Welcome to the app!")
        binding.progressBar.gone()
        hideKeyboard(binding.root)
    }
}
```

---

## Literals and Constants

### Boolean Literals

```kotlin
val isEnabled: Boolean = true
val isDisabled: Boolean = false

// Mobile usage
if (BuildConfig.DEBUG) {
    Log.d(TAG, "Debug mode enabled")
}
```

### Numeric Literals

```kotlin
// Integer literals
val timeout = 5000
val maxRetries = 3
val hexColor = 0xFF0000
val binaryFlags = 0b1010

// Long literals
val timestamp = 1234567890L
val fileSize = 1024L * 1024L * 10L // 10 MB

// Unsigned literals
val unsignedInt = 42u
val unsignedLong = 42uL

// Float literals
val progress = 0.75f
val animationDuration = 300.0f

// Double literals
val latitude = 37.7749
val longitude = -122.4194
```

### Character Literals

```kotlin
val newline = '\n'
val tab = '\t'
val quote = '\''
val unicode = '\u0041' // 'A'
```

### String Literals

```kotlin
// Single-line strings
val message = "Hello, World!"
val path = "C:\\Users\\Documents"

// Multi-line strings
val jsonTemplate = """
    {
        "user_id": "$userId",
        "timestamp": ${System.currentTimeMillis()},
        "data": {
            "key": "value"
        }
    }
""".trimIndent()

// String templates
val welcomeMessage = "Welcome, ${user.name}!"
val progressText = "Progress: ${(current * 100 / total).toInt()}%"
```

### Null Literal

```kotlin
var user: User? = null
val defaultConfig: Config? = null

// Mobile usage
fun loadUserProfile(userId: String?): User? {
    return if (userId != null) {
        userRepository.getUser(userId)
    } else null
}
```

---

## Grammar Rules

### Function Declaration

```kotlin
// Basic function
fun functionName(parameter: Type): ReturnType {
    return value
}

// Mobile example
fun validateEmail(email: String): Boolean {
    return android.util.Patterns.EMAIL_ADDRESS.matcher(email).matches()
}

// Extension function
fun String.isValidEmail(): Boolean {
    return android.util.Patterns.EMAIL_ADDRESS.matcher(this).matches()
}

// Suspending function
suspend fun fetchUserData(userId: String): User {
    return apiService.getUser(userId)
}
```

### Class Declaration

```kotlin
// Basic class
class ClassName(parameter: Type) {
    // Class body
}

// Android component
class UserProfileFragment : Fragment() {
    companion object {
        fun newInstance(userId: String) = UserProfileFragment().apply {
            arguments = bundleOf("user_id" to userId)
        }
    }
}

// Data class
data class User(
    val id: String,
    val name: String,
    val email: String?
)

// Sealed class
sealed class NetworkResult<out T> {
    data class Success<T>(val data: T) : NetworkResult<T>()
    data class Error(val exception: Throwable) : NetworkResult<Nothing>()
    object Loading : NetworkResult<Nothing>()
}
```

### Property Declaration

```kotlin
// Read-only property
val property: Type = value

// Mutable property
var property: Type = value

// Property with custom getter/setter
var property: Type
    get() = field
    set(value) {
        field = value
        notifyPropertyChanged()
    }

// Delegated property
val lazyProperty: Type by lazy { computeValue() }
```

### Type Constraints

```kotlin
// Single constraint
fun <T : Comparable<T>> sort(list: List<T>): List<T> {
    return list.sorted()
}

// Multiple constraints
fun <T> saveToDatabase(item: T): Long
    where T : Serializable,
          T : DatabaseEntity {
    return database.insert(item)
}
```

### Import Statements

```kotlin
// Regular import
import com.example.User

// Import with alias
import com.example.User as ExampleUser

// Import all
import com.example.utils.*

// Mobile-specific imports
import androidx.fragment.app.Fragment
import androidx.lifecycle.ViewModel
import androidx.navigation.fragment.navArgs
```

---

## Best Practices for Mobile Development

### 1. Use Appropriate Visibility Modifiers

```kotlin
class UserRepository {
    // Public API
    suspend fun getUser(id: String): User = internalGetUser(id)

    // Internal implementation details
    private suspend fun internalGetUser(id: String): User {
        return apiService.fetchUser(id)
    }

    // Protected for inheritance
    protected open fun validateUserId(id: String): Boolean {
        return id.isNotEmpty()
    }
}
```

### 2. Leverage Null Safety

```kotlin
// Good: Safe call chains
fun displayUserInfo(user: User?) {
    user?.profile?.avatar?.let { url ->
        imageLoader.load(url, imageView)
    }

    val displayName = user?.name ?: "Guest"
    textView.text = displayName
}

// Avoid: Excessive null assertions
// Bad: user!!.name!!.length (risky)
```

### 3. Use Appropriate Keywords for Android Components

```kotlin
// Proper lifecycle handling
abstract class BaseFragment : Fragment() {
    // Template method pattern
    final override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        setupViews()
        observeData()
    }

    protected abstract fun setupViews()
    protected abstract fun observeData()
}

// Sealed classes for state management
sealed class UiState {
    object Loading : UiState()
    data class Success(val data: List<Item>) : UiState()
    data class Error(val message: String) : UiState()
}
```

### 4. Effective Use of Coroutines Keywords

```kotlin
class NetworkManager {
    // Suspending functions for async operations
    suspend fun fetchData(): String = withContext(Dispatchers.IO) {
        // Network call
        apiService.getData()
    }

    // Inline functions for performance
    inline fun <T> measureTime(operation: () -> T): T {
        val start = System.currentTimeMillis()
        val result = operation()
        Log.d(TAG, "Operation took ${System.currentTimeMillis() - start}ms")
        return result
    }
}
```

This comprehensive reference covers all essential Kotlin keywords and syntax with practical mobile development examples, helping developers understand both the language fundamentals and their application in Android development contexts.