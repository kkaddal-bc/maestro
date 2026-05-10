# Scan Patterns — Kotlin

Framework-specific patterns for extracting API contracts and DB schema from Kotlin repositories.

---

## HTTP — Spring Boot

**Controller class** — `@RestController` or `@Controller` on the class. Base path from `@RequestMapping("/base")` on the class.

**Method-level routes** — combine class base path with method annotation path:
```kotlin
@GetMapping("/orders/{id}")
@PostMapping
@PutMapping("/{id}")
@DeleteMapping("/{id}")
@PatchMapping("/{id}")
@RequestMapping(method = [RequestMethod.GET], path = ["/orders"])
```

**Auth** — look for:
- `@PreAuthorize("hasRole('ADMIN')")` → `🔒 Role: ADMIN`
- `@Secured("ROLE_USER")` → `🔒 Role: USER`
- `@AuthenticationPrincipal` parameter → `🔒 JWT` (authenticated user injected)
- Spring Security config (`SecurityFilterChain`) for route-level rules — check for `permitAll()` → `public`

**Request DTO** — parameter with `@RequestBody`: `fun create(@RequestBody body: CreateOrderRequest)`  
**Path params** — `@PathVariable id: UUID`  
**Query params** — `@RequestParam page: Int = 0`  
**Response type** — return type annotation or `ResponseEntity<OrderResponse>`

```kotlin
@RestController
@RequestMapping("/api/v1/orders")
class OrderController {
    @PostMapping
    fun create(@RequestBody body: CreateOrderRequest): ResponseEntity<OrderResponse>

    @GetMapping("/{id}")
    fun getById(@PathVariable id: UUID): OrderResponse
}
```

## HTTP — Ktor

Routes defined with routing DSL — look for nested `route`, `get`, `post`, `put`, `delete`, `patch` blocks:

```kotlin
routing {
    route("/api/v1") {
        route("/orders") {
            post { val body = call.receive<CreateOrderRequest>() }
            get("/{id}") { val id = call.parameters["id"] }
        }
    }
}
```

**Request body** — `call.receive<DtoClass>()`  
**Path params** — `call.parameters["name"]`  
**Query params** — `call.request.queryParameters["page"]`  
**Auth** — `authenticate("jwt") { ... }` block wrapping the route → `🔒 JWT`

---

## WebSocket — Spring

**Handler class** — `@Controller` with `@MessageMapping`:

```kotlin
@Controller
class TradingWsController {
    @MessageMapping("/order.create")        // incoming event
    @SendTo("/topic/orders")               // broadcast destination
    fun handleCreate(payload: CreateOrderPayload): OrderEvent

    @SubscribeMapping("/positions")         // on subscribe
    fun onSubscribe(): List<Position>
}
```

**Destination prefix** — check `WebSocketMessageBrokerConfigurer.configureMessageBroker` for app and broker prefixes.

## WebSocket — Ktor

```kotlin
webSocket("/ws/trading") {
    for (frame in incoming) {
        val msg = Json.decodeFromString<WsMessage>(frame.readText())
        // ...
        send(Json.encodeToString(WsResponse(...)))
    }
}
```

Look for `send(...)` and `receive<...>()` / `frame.readText()` calls for the message shapes.

---

## gRPC — Proto files

Read `.proto` files directly. Extract:
```proto
service OrderService {
  rpc CreateOrder (CreateOrderRequest) returns (OrderResponse);
  rpc WatchOrders (WatchRequest) returns (stream OrderEvent);
}
message CreateOrderRequest {
  string symbol = 1;
  double quantity = 2;
  OrderSide side = 3;
}
```
- `stream` keyword → streaming side
- `enum` types → note allowed values

## gRPC — Kotlin service implementation

```kotlin
@GrpcService
class OrderGrpcService : OrderServiceGrpcKt.OrderServiceCoroutineImplBase() {
    override suspend fun createOrder(request: CreateOrderRequest): OrderResponse
    override fun watchOrders(request: WatchRequest): Flow<OrderEvent>
}
```

Cross-reference the `.proto` for message field definitions — the Kotlin stubs are generated and don't add info.

---

## Database — JPA / Hibernate

**Table name** — `@Entity @Table(name = "orders")` on the class, or class name snake_cased if no `@Table`.  
**Columns** — `@Column(name = "user_id", nullable = false, unique = true, length = 255)`.  
**Primary key** — `@Id @GeneratedValue(strategy = GenerationType.IDENTITY)` or `UUID`.  
**Relations:**
```kotlin
@ManyToOne(fetch = FetchType.LAZY)
@JoinColumn(name = "user_id", nullable = false)
val user: UserEntity                          // FK lives here

@OneToMany(mappedBy = "user", cascade = [CascadeType.ALL])
val orders: List<OrderEntity>                 // inverse side

@ManyToMany
@JoinTable(name = "user_roles",
    joinColumns = [JoinColumn(name = "user_id")],
    inverseJoinColumns = [JoinColumn(name = "role_id")])
val roles: Set<RoleEntity>
```
**Indexes** — `@Table(indexes = [Index(name = "idx_email", columnList = "email", unique = true)])`.  
**Audit fields** — `@CreatedDate`, `@LastModifiedDate` from Spring Data.

## Database — Exposed

```kotlin
object Orders : UUIDTable("orders") {
    val userId    = uuid("user_id").references(Users.id)
    val symbol    = varchar("symbol", 20)
    val quantity  = decimal("quantity", 18, 8)
    val status    = enumerationByName("status", 20, OrderStatus::class)
    val createdAt = datetime("created_at").defaultExpression(CurrentDateTime)
}
```
- Column type method name → SQL type (e.g. `varchar(col, 255)` → `VARCHAR(255)`)
- `.references(Table.col)` → FK
- `.uniqueIndex()`, `.index()` → index

## Database — Flyway Migrations

Read SQL files in version order (`V1__`, `V2__`, ...). Extract:
- `CREATE TABLE` — columns, types, constraints
- `ALTER TABLE ADD COLUMN` — additions not yet in entity classes
- `CREATE INDEX` / `CREATE UNIQUE INDEX`
- `ALTER TABLE ADD CONSTRAINT ... FOREIGN KEY`

These are the ground truth for the actual DB state, especially for columns added after the entity was first written.

---

## DTOs — Kotlin data classes

```kotlin
data class CreateOrderRequest(
    @field:NotBlank val symbol: String,
    @field:NotNull @field:Positive val quantity: BigDecimal,
    @field:NotNull val side: OrderSide,
    @field:Min(0) val leverage: Int = 1,
    val clientOrderId: String? = null       // nullable = optional
)
```

For each field extract:
- **name**: property name
- **type**: Kotlin type (note `?` = nullable/optional)
- **validation**: Jakarta/javax annotations (`@NotNull`, `@NotBlank`, `@Size`, `@Min`, `@Max`, `@Email`, `@Pattern`, `@Positive`)
- **required**: non-nullable without default = required; nullable (`?`) or has default = optional
- **Jackson**: `@JsonProperty("snake_case")`, `@JsonIgnore`, `@JsonInclude`

Sealed classes used as discriminated unions — document each subclass as a variant.
