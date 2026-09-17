Problem Statement: Library Book Management System
Objective:
Design and implement a backend service that allows users to manage books, borrow and return them, and track availability.

Functional Requirements: 
Book Management
Add a new book to the library.
Retrieve book details by ID.
List books with filters (e.g., available books, books by author).
User Management
Register a new user.
Fetch user details.
Borrow & Return Books
Borrow a book if it is available.
Return a borrowed book, updating its availability.
List books borrowed by a user.

Non-Functional Requirements
Expose a RESTful API with appropriate HTTP methods.
Optimize data retrieval and updates.
Follow clean code principles with modular design.
Write basic unit tests for key functionalities.
 
Constraints:
A book can be borrowed by one user at a time.
Users can borrow up to 3 books at once.

Entities:

User:
- ID
- Name

Book
- ID
- Name
- AvailabilityQuantity (MutexLocks)

Author
- ID
- Name

AuthorBookMapping
- ID
- BookID
- AuthorID

LibraryLedger
- ID
- UserID
- ActionType (Borrow/Return) - Lock
- BookId
- ReturnedAt
- UpdatedAt
- CreatedAt


<!--Borrow a same copy book-->

U1: Borrow(ID1) Borrow(ID2)
<!--U2: Borrow(ID1)-->

Borrow(ID1) {
  <!--Checking the idempotency key-->
  Single Transaction {
  UserMutexLock ()
  Get availability of book
  Get count(book) from LibraryLedger where ActionType = Borrow - > 2, 2 - CountChecker
   -> Mutex Lock for Book Row with ID1
      Book Availability Update - On conflict cancel
   -> Release lock
   Library Ledger Addition
   <!--CountChecker -> Check for less than 3 -> Else Revert-->
   commit;
  }

  <!--CountChecker -> Check for less than 3 -> Else Revert-->
}
