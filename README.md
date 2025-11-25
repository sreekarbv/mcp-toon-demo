# mcp-toon-demo

## Build steps
cd server
go build -o server .

cd ../client
go build -o client .


## Example RUN
./clinet
Enter your question:
fetch me number users from India
JSON Payload length  59589
TOON Payload length  27610

--- Token & Cost Analysis (GPT-4o) ---
TOON: 10837 tokens  →  $0.054185
JSON: 18029 tokens  →  $0.090145
Savings: 7192 tokens (39.89%)
--------------------------------------

=== GPT-4o Answer ===
The cities in India, as per the user data, are Mumbai, Pune, and Bangalore. Let's calculate the number of users from each of these cities and sum them up to get the total number of users from India:

- Mumbai: 64 users
- Pune: 145 users
- Bangalore: 134 users

Total users from India = 64 + 145 + 134 = 343 users.
