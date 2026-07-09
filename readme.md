The Requirements:

CreateWallet(userId string, initialBalance int): Initializes a wallet.

Transfer(fromUserId string, toUserId string, amount int): Moves money. It must fail gracefully if the fromUserId has insufficient funds.

GetStatement(userId string): Returns a list of all transactions (credits and debits) for that specific user, sorted by the most recent.
