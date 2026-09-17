## Hierarchical Restaurant Menu LLD

This problem is really three questions tied together:

1. How do we model the menu domain cleanly?
2. How do we calculate dynamic cart pricing without hardcoding rules everywhere?
3. How do we serve this nested menu fast at scale?

The strongest answer is the one where all three parts connect. The domain model should naturally support pricing, and the pricing model should naturally support caching and scale.

---

## 1. How to Think About This Problem

When solving LLDs like this, use this order:

### Step 1: Identify the user action, not just the entities

The real user flow is:

- user opens menu
- user picks an item
- user selects one option from each required variant group
- user selects add-ons
- system validates the configuration
- system calculates final price
- system stores the customized cart item

That means this is not just a static entity-modeling problem. It is a:

- catalog modeling problem
- configuration validation problem
- pricing problem
- scale/read optimization problem

If you think only in terms of getters/setters, you will miss the important behavior.

### Step 2: Separate stable catalog data from user selection data

You need two kinds of objects:

- Catalog side:
  menu definitions, items, variant groups, add-on groups, pricing rules
- Cart side:
  what the user actually selected for one customized item

This separation is one of the biggest design decisions in the problem.

If catalog objects also try to represent user selections, the model becomes confusing very quickly.

### Step 3: Identify the source of pricing complexity

The tricky part is this line:

"the price of an add-on can dynamically scale based on the parent variant chosen"

That means add-on price is not always absolute. It may depend on:

- selected size
- selected crust
- selected variant combination

So pricing should not live only as a single `PriceDelta` field everywhere. It needs a rule-based or strategy-based representation.

### Step 4: Decide where behavior should live

For LLD, ask:

- where should validation live?
- where should pricing live?
- where should rule resolution live?

Good answer:

- entities hold structure and invariants
- service layer orchestrates use cases
- pricing rules are encapsulated behind interfaces

### Step 5: Only then think about patterns

Many people jump to "Decorator" too early. Pattern selection should come after understanding the volatility:

- menu structure changes
- pricing rules change
- read traffic is huge

So choose patterns that isolate change.

---

## 2. Best Design Approach

### Core domain split

Use two major aggregates:

1. `Catalog`
2. `CartItem`

The catalog is what the restaurant publishes.
The cart item is a concrete user customization.

### Recommended model

```go
type Menu struct {
    ID            string
    RestaurantID  string
    Sections      []MenuSection
    Version       int
}

type MenuSection struct {
    ID       string
    Name     string
    ItemIDs   []string
}

type CatalogItem struct {
    ID                string
    Name              string
    Description       string
    BasePrice         Money
    VariantGroups     []VariantGroup
    AddOnGroups       []AddOnGroup
}

type VariantGroup struct {
    ID          string
    Name        string
    Required    bool
    MinSelect   int
    MaxSelect   int
    Options     []VariantOption
}

type VariantOption struct {
    ID          string
    Name        string
    PriceDelta  Money
    AddOnGroups []AddOnGroup
}

type AddOnGroup struct {
    ID           string
    Name         string
    MinSelect    int
    MaxSelect    int
    Choices      []AddOnChoice
}

type AddOnChoice struct {
    ID          string
    Name        string
    ItemRefID   string
    PricingRule PricingRule
}
```

Then create selection-side models:

```go
type CartItem struct {
    CatalogItemID        string
    SelectedVariants     []SelectedVariant
    SelectedAddOns       []SelectedAddOn
    Quantity             int
}

type SelectedVariant struct {
    VariantGroupID   string
    VariantOptionID  string
}

type SelectedAddOn struct {
    AddOnGroupID   string
    AddOnChoiceID  string
    Quantity       int
}
```

This is much cleaner because:

- catalog remains reusable
- cart item remains lightweight
- price calculation can use catalog + selection together

---

## 3. Best Pattern for Pricing

### Best answer: Strategy + Composite, not Decorator alone

If asked for one pattern, I would say:

`Strategy pattern` for pricing rules, combined with a small `Composite-style` price breakdown.

Decorator can work for accumulating price components, but by itself it is not the best fit for dynamic rule resolution based on selected parent variant. Strategy handles rule variation more cleanly.

### Why Strategy fits better

Different add-ons may price differently:

- fixed price: extra cheese is always +40
- variant-relative: large pizza topping is +60, medium is +40
- percentage based: add 10 percent of selected base
- free for some variants, paid for others

These are different algorithms. That is classic Strategy.

### Pricing interface

```go
type PricingRule interface {
    Price(ctx PricingContext) Money
}
```

`PricingContext` can contain:

- base catalog item
- selected variants
- selected add-on
- resolved parent variant
- quantity

Example strategies:

```go
type FixedPriceRule struct {
    Amount Money
}

type VariantBasedPriceRule struct {
    PriceByVariantOptionID map[string]Money
}

type PercentagePriceRule struct {
    Percent float64
}
```

### Where Composite helps

For maintainability, calculate a price breakdown as components:

- base price
- variant deltas
- add-on charges
- taxes or discounts if needed later

That can be represented as:

```go
type PriceComponent struct {
    Code   string
    Amount Money
}

type PriceBreakdown struct {
    Components []PriceComponent
}
```

Then total is the sum of components.

This is interview-friendly because it shows extensibility without overengineering.

---

## 4. How Price Calculation Should Flow

### PriceCalculatorService responsibilities

The `PriceCalculatorService` should:

1. load the catalog item
2. validate selected variants
3. validate selected add-ons against allowed groups
4. build pricing context
5. calculate base + variant deltas + add-ons
6. return structured breakdown and total

### Pseudocode

```go
func (s *PriceCalculatorService) Calculate(cartItem CartItem) (PriceBreakdown, error) {
    item := s.catalogRepo.GetItem(cartItem.CatalogItemID)

    err := s.validator.Validate(item, cartItem)
    if err != nil {
        return PriceBreakdown{}, err
    }

    breakdown := NewPriceBreakdown()

    breakdown.Add("base_price", item.BasePrice)

    selectedVariants := resolveSelectedVariants(item, cartItem)
    for _, v := range selectedVariants {
        breakdown.Add("variant_delta", v.PriceDelta)
    }

    ctx := PricingContext{
        Item:             item,
        SelectedVariants: selectedVariants,
        CartItem:         cartItem,
    }

    for _, addOn := range resolveSelectedAddOns(item, cartItem) {
        amount := addOn.PricingRule.Price(ctx.WithAddOn(addOn))
        breakdown.Add("addon_price", amount)
    }

    return breakdown, nil
}
```

### Important interview point

Validation should happen before pricing.

Example:

- medium pizza may allow 2 toppings
- large pizza may allow 4 toppings
- thin crust may disable cheese burst add-on

These are configuration rules, not just price rules.

So have a separate `CartItemValidator`.

---

## 5. Modeling Dynamic Add-On Pricing Correctly

This is the heart of the problem.

There are two clean ways to model it.

### Option A: Pricing strategy on each add-on choice

```go
type AddOnChoice struct {
    ID          string
    Name        string
    PricingRule PricingRule
}
```

This is the best design if pricing logic may vary a lot.

### Option B: Variant-specific price map

```go
type AddOnChoice struct {
    ID                     string
    Name                   string
    DefaultPriceDelta      Money
    PriceByVariantOptionID map[string]Money
}
```

This is simpler and easier for interviews if the scope is limited.

### Which one should you say in interview?

Say:

- for a simple system, use `PriceByVariantOptionID`
- for an extensible production design, wrap pricing in `PricingRule`

That shows good judgment.

---

## 6. Evaluation of Your Partial Solution

Your current model has a good starting instinct:

- you identified the main entities
- you captured that variant options can have their own add-on categories
- you introduced price delta for variants and add-ons

Those are all useful observations.

### What is good

1. `CatalogItem -> VariantGroups -> VariantOptions` is a good backbone.
2. Attaching `AddonCategory` to `VariantOption` is directionally correct because available add-ons may depend on selected variant.
3. Constructor methods are a reasonable start for keeping object creation consistent.

### What is missing or weak

1. There is no cart-side model.

Without `CartItem`, `SelectedVariant`, and `SelectedAddOn`, you cannot cleanly represent a customer’s customization.

2. `AddonChoice` is under-modeled.

Right now:

```go
type AddonChoice struct {
    ID         uuid.UUID
    PriceDelta float64
}
```

This is not enough. It likely needs:

- name
- referenced catalog item or SKU
- pricing rule or price map
- availability flags

3. `CatalogItem` and `VariantOption` both directly own `AddonCategory`, but the selection rules are not explicit.

You need clarity around:

- are item-level add-ons always available?
- does variant-level add-on override item-level add-on?
- do they merge?

4. `float64` should not be used for price.

Money should be represented as:

- integer minor units like paise/cents, or
- dedicated `Money` type

Using `float64` for currency causes precision issues.

5. Missing validation constraints on variant groups.

You have `MinSelection` and `MaxSelection` for add-ons, but not for variants.
That means you cannot express:

- exactly one size required
- up to two crust choices

6. No behavior yet for price calculation.

The entities are only structure right now. That is okay as a first pass, but the problem’s main challenge is in runtime pricing and validation.

7. Naming can be made more explicit.

`AddonCategory` is okay, but `AddOnGroup` is more standard and communicates selection semantics more clearly.

8. `Menu` currently has `[]CatalogItem`, but real menus often have sections.

A `MenuSection` helps both domain clarity and read-path performance.

### Overall evaluation

Your current solution is a decent entity sketch, but it is not yet a complete LLD answer.

If I were grading it in an interview:

- entity identification: good
- behavioral modeling: incomplete
- pattern selection: not yet expressed
- scalability thinking: not yet modeled

So the base is workable, but the strongest next step is to introduce:

- selection-side entities
- validator service
- pricing abstraction
- menu read model

---

## 7. Suggested Improved Go Design

If you continue coding this solution, I would structure packages like this:

```text
menu/
  entity.go
  pricing.go
  validator.go
  repository.go
  service.go
  readmodel.go
```

### `entity.go`

Keep pure domain models:

- Menu
- MenuSection
- CatalogItem
- VariantGroup
- VariantOption
- AddOnGroup
- AddOnChoice
- CartItem

### `pricing.go`

Keep:

- `PricingRule` interface
- concrete strategies
- `PricingContext`
- `PriceBreakdown`

### `validator.go`

Keep:

- required variant validation
- min/max add-on validation
- variant-add-on compatibility validation

### `service.go`

Keep:

- `PriceCalculatorService`
- orchestration logic only

This separation will make your code much easier to explain.

---

## 8. Read Path Design for Millions of Users

This is the system design part hidden inside an LLD question.

The main requirement is:

"avoid complex SQL joins on every page load"

That means your serving path should not assemble the menu from normalized tables on every request.

### Recommended read path

Use a precomputed denormalized menu document.

Flow:

1. restaurant updates menu in write store
2. system publishes menu-changed event
3. menu materializer builds a denormalized read model
4. read model is stored in cache and document store
5. menu page reads the prebuilt document directly

### Storage design

Write side:

- normalized relational tables for integrity and editing

Read side:

- denormalized JSON document per restaurant menu version
- stored in Redis and/or document DB

### Example read model key

```text
menu:{restaurant_id}:{menu_version}
```

### Why this works

- one lookup instead of many joins
- good cache locality
- easy invalidation by version bump
- easy horizontal scaling

---

## 9. Caching Strategy

### L1 cache

- in-process cache on API servers
- short TTL
- useful for hottest menus

### L2 cache

- Redis or Memcached
- stores full denormalized menu document

### Persistent read model

- document DB or object store
- source of truth for read path if Redis misses

### Cache flow

1. request comes for restaurant menu
2. check local in-memory cache
3. if miss, check Redis
4. if miss, load denormalized document from read store
5. repopulate Redis and local cache

### Invalidation strategy

Best answer:

- versioned menu documents
- publish invalidation on menu update
- old versions naturally expire

This is better than hard deleting one mutable cache key because versioning avoids race conditions.

### What not to do

Avoid:

- rebuilding nested menu from SQL joins on every request
- cache entry per tiny sub-entity for read path
- synchronous cache rebuild during user request

---

## 10. Best High-Level Interview Answer

If asked verbally, a strong concise answer would be:

"I would separate catalog modeling from cart customization. On the catalog side I’d model `CatalogItem`, `VariantGroup`, `VariantOption`, `AddOnGroup`, and `AddOnChoice`. On the cart side I’d model the user’s selected variants and add-ons in a `CartItem`. For pricing, I’d use a `PricingRule` strategy per add-on choice so the final price can depend on the selected parent variant. The `PriceCalculatorService` would first validate the configuration, then compute a structured price breakdown from base price, variant deltas, and add-on rule evaluation. For scale, I’d keep normalized data on the write path but build a denormalized versioned menu document for the read path, cached aggressively in Redis and optionally in-process, so page loads avoid expensive joins." 

That answer is compact but hits all three parts.

---

## 11. Things to Keep in Mind in Future LLD Problems

### Always separate these four concerns

1. domain entities
2. user actions/use cases
3. rule evaluation
4. storage/read optimization

### Ask these questions early

1. Which objects are master data and which are transaction data?
2. Which rules are likely to change often?
3. What needs validation before processing?
4. Which path is read-heavy and needs denormalization?
5. What should be extensible without changing existing classes?

### Common mistakes to avoid

1. Only drawing entities and forgetting behavior.
2. Putting pricing logic inside random entity getters.
3. Mixing catalog definition with cart selection state.
4. Using `float64` for money.
5. Ignoring validation rules.
6. Giving a normalized DB answer for a read-heavy nested UI problem.

---

## 12. If You Want to Improve Your Current Code Next

The best next implementation steps are:

1. introduce `Money` type
2. add cart-side entities
3. add min/max constraints to variant groups
4. define `PricingRule` interface
5. implement `PriceCalculatorService`
6. implement `CartItemValidator`
7. create a denormalized menu response model

If you do just these, your solution will move from "entity sketch" to "complete LLD answer".
