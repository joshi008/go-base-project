//  Splitwise Design
//
// Functional Requirement
// User should be able to add Split on them
// Split should be
// 		Equal, Percentage and Exact
// User should be able to check their current balance / owe list
// 		How much they get from others list

// Non Functional Requirements
// Consistency > Availability
// We should be able to have 1M daily active users doing 1/5th adding transaction

// Entities
//
//
// User
// - ID
// - Name
// - Expenses[]
//

1 -> N Browers
1 <- N Lenders

UserLevelBalancingSheet
- OwnerUserId
- TransactingUserId
- Amount

//
// Expense
// - ID
// - Amount
// - LenderUserId
// - Split Type (Equal, Percentage and Exact)
// - Splits[]
//
// Splits
// - ID
// - ExpenseId
// - BrowerUserId
// - Amount
// - LenderUserId
//
// APIs
// /Post -> Adding Expense
// /Get -> Checking all the transactions
//  Req: UserId
//  Select DISTINCT(ExpenseId) Splits s where s.UserId= UserId and s.LenderUserId=UserId    AS E;
// SELECT * from EXPENSES where id in E;
//
// /Get -> User Wise Balances
// Request: UserId
// SELECT * from UserLevelBalancingSheet where
//
// /POST -> Settle Balance
//
