Machine Coding Round – Eligibility Rule Engine
Problem Statement
Design and implement an in-memory Eligibility Rule Engine that evaluates whether a user is eligible for a feature, based on configurable rules.
Rules
A rule consists of one or more conditions on user attributes. Conditions can be combined using logical operators (AND, OR, NOT), and combinations can themselves be nested inside other combinations to arbitrary depth.
Example rules:
(country = "IN" AND kycVerified = true)

 (country = "IN" AND appVersion >= "2.0.0") OR (monthlyVolume > 100000)
 

 NOT(isBlocked = true) AND (accountAgeInDays > 30)
User Context
Each evaluation is performed against a user context — a set of attributes describing the user:
{
   "userId": "123",
   "country": "IN",
   "kycVerified": true,
   "monthlyVolume": 250000,
   "accountAgeInDays": 180,
   "isBlocked": false,
   "appVersion": "2.5.0"
 }
Requirements
Support defining a rule for a feature as one or more conditions, combined using logical operators (AND, OR, NOT), nestable to arbitrary depth.
Support evaluating a rule against a given user context, returning a boolean eligibility result.
Implement exactly these comparison operators — no others are required for this round: =, !=, >, <, >=, <=
Attribute types are limited to: String (e.g. country), Boolean (e.g. kycVerified), Number (e.g. monthlyVolume, accountAgeInDays), and version strings (e.g. appVersion, compared numerically by segment — "2.10.0" > "2.9.0"). No other types need to be supported.
Define behavior for an attribute that is missing from the user context (your choice — fail-closed, throw, or skip — but be explicit and consistent).
You should be able to add new operators or new condition types with minimal changes to existing code.
Constraints
In-memory implementation only — no persistence or external systems.
Rules are constructed programmatically (via code/builder), not parsed from a string.
Focus on clean design, correctness, extensibility, and maintainability over feature breadth.
Deliverables
Working code that builds sample rules (including the examples above) and evaluates them against sample user contexts, including at least one case where an attribute is missing.
A short note (comments or README) on how the design would accommodate a new operator or condition type.
