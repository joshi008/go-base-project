We are building a dashboard to show a single number - the total cost of all deliveries - on screens in the accounting department offices.

At first, we want the following functions:

`AddDriver(driverId [integer], usdHourlyRatePerDelivery [float])`

The given driver will not already be in the system
The hourly rate applies per delivery, so a driver can be paid multiples of this rate per hour for simultaneous deliveries


`RecordDelivery(driverId [integer], startTime, endTime)`

Discuss the time format you choose
Times require minimum one-second precision
The given driver will already be in the system
All deliveries will be recorded immediately after the delivery is completed
No delivery will exceed 3 hours


`GetTotalCost() -> string`

Return the total, aggregated cost of all drivers' deliveries recorded in the system, to 2 decimal places
For example, return "135.30" if one driver is in the system and has a total cost of 100.30 USD and another is in the system and has a total cost of 35.00 USD.
This will be used for a live dashboard




Q2 :
The accounting department now wants to use the live dashboard you built to see how much money is owed in total to all drivers.

Add the following functions:

`PayUpTo (payTime)`

Pay all drivers for recorded deliveries which ended up to and including the given time
payTime shall not exceed the current system time


`GetTotalCostUnpaid() -> string`

Return the total, aggregated cost of all drivers' deliveries which have not been paid
