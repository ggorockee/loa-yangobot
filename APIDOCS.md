# GETTING STARTED

GETTING STARTED
You can start using the Lostark Open API by following basic procedures below to understand how it works.

Login
Before using any of the Lostark Open APIs, you need to sign in to your Stove.com account. If you don't have one yet, please sign up here. There is no additional step required to sign up for the Lostark Open API website; signing up for Stove will suffice. Once you sign in, you can create a new client immediately.

Create A Client
Follow these steps to create a client:


1. Click CREATE A NEW CLIENT button or this link
2. In the Client Name field, enter a name to identify your client in the list view. This name will only be associated with your account.
3. In the Client URL field, enter the URL of the service that will be using this client.
4. In the Client Description field, describe what your application will do.
5. Read and agree Terms of Use and Privacy Policy. ( These are NOT optional )
6. After creating your client, you can view the details on the MY CLIENTS page or in the popup that appears when you click the red identity button in the top right corner.
JWT Key
When you create a client, a JWT is issued immediately. We have an OAuth security layer in place, but you don't need to worry about refreshing expired tokens or performing tedious tasks to request a new access token.

The issued security token will remain valid indefinitely unless we determine that your key is not secure or the user deletes the client. We strongly recommend storing the JWT token in a secure place, such as a trusted server-side application.

For example, if you receive a token called "abcdefghijklmnopqrstuvwxyz", you should set the authorization header as follows:

Example	Validity
"Authorization: bearer abcdefghijklmnopqrstuvwxyz"	O => VALID
"Authorization: abcdefghijklmnopqrstuvwxyz"	X => INVALID (missing "bearer")
"Authorization: bearer {abcdefghijklmnopqrstuvwxyz}"	X => INVALID
"Authorization: bearer <abcdefghijklmnopqrstuvwxyz>"	X => INVALID
"Authorization: bearer{abcdefghijklmnopqrstuvwxyz}"	X => INVALID
"Authorization: bearerabcdefghijklmnopqrstuvwxyz"	X => INVALID
API Documentation
The Lostark Open API documentation allows you to interact with the API's resources without actual implementation logic. You can explore what we offer and make actual requests using your JWT. Always ensure you click the "Try it out" button to activate the "Execute" button, which finalizes the request setup, calls the API, and provides the result.

Exploring the request parameters and checking the responses before implementing the Lostark Open APIs will save you a significant amount of time and effort.

Swagger Authorization
To authorize and call an API in the Lostark Open API documentation, you need to pass your JWT as follows:


1. Click AUTHORIZE button
Authorize
2. Enter your JWT in the VALUE field
※ Make sure to avoid the bad examples presented above for successful validations of your token.
Available authorizations
apiKeyAuth (apiKey)
Enter your API JWT here to be used with the interactive documentation
ex:) bearer JWT
※ if unauthorized with your valid token, please review the bad examples.
※ Ensure you avoid the invalid cases.

Name: authorization

In: header

Value:
bearer your_JWT_token
Authorize
Close
Throttling
Clients are limited to 100 requests per minute. Exceeding this quota results in a 429 response until the quota is reset. The quota is automatically renewed every minute, so your application should resume working after a minute once the limit is reached.


Throttling metadata in the response header
429 responses are probably meaningless for you and you may feel frustrated. Throttle your client application then! these metadata below will come in handy for you.
X-RateLimit-Limit : Request per minute (int)
X-RateLimit-Remaining : The number of available requests for your client. (int)
X-RateLimit-Reset : The epoch time for the next quota refresh. (long)

# USAGE GUIDE
This guide describes how to interact with the Lostark Open APIs. It is intended for software developers and walks you through implementation details and sample code.

Basic HTTP
Making a valid HTTP request to the Lostark Open API services involves three parts:


1. Set the correct HTTP verb.
2. Set application/json in the 'accept' header and if required, the 'content-type' header as well.
3. Set the bearer token in the authorization header

Remember, your JWT must always be included with every API request. Specify it in the authorization header.


curl example
curl -X 'GET' 'https://developer-lostark.game.onstove.com/exmaple/api' 
-H "accept: application/json" 
-H "authorization: bearer your_JWT"

Javascript example
var xmlHttpRequest = new XMLHttpRequest();
xmlHttpRequest.open("GET", "https://developer-lostark.game.onstove.com/exmaple/api", true);
xmlHttpRequest.setRequestHeader('accept', 'application/json');
xmlHttpRequest.setRequestHeader('authorization', 'bearer your_JWT');
xmlHttpRequest.onreadystatechange = () => { };
xmlHttpRequest.send();

The Lostark Open API GET Request Example
GET /guilds/rankings API takes a 'serverName' string as a request parameter. Since it is a query string, compose your request URL as follows:


Query string parameter curl example
curl -X 'GET' 'https://developer-lostark.game.onstove.com/guilds/rankings?serverName=%EB%A3%A8%ED%8E%98%EC%98%A8'
-H 'accept: application/json'
-H 'authorization: bearer your_JWT
Query string parameter Javascript example
var xmlHttpRequest = new XMLHttpRequest();
xmlHttpRequest.open("GET", "https://developer-lostark.game.onstove.com/guilds/rankings?serverName=%EB%A3%A8%ED%8E%98%EC%98%A8", true);
xmlHttpRequest.setRequestHeader('accept', 'application/json');
xmlHttpRequest.setRequestHeader('authorization', 'bearer your_JWT');
xmlHttpRequest.onreadystatechange = () => { };
xmlHttpRequest.send();

GET /armories/characters/{characterName}/profiles API takes a 'characterName' string as a request parameter. Since it is a path string, you need to compose your request URL like this below


Path parameter curl example
curl -X 'GET' 'https://developer-lostark.game.onstove.com/armories/characters/coolguy/profiles'
-H 'accept: application/json'
-H 'authorization: bearer your_JWT'
Path parameter Javascript example
var xmlHttpRequest = new XMLHttpRequest();
xmlHttpRequest.open("GET", "https://developer-lostark.game.onstove.com/armories/characters/coolguy/profiles", true);
xmlHttpRequest.setRequestHeader('accept', 'application/json');
xmlHttpRequest.setRequestHeader('authorization', 'bearer your_JWT');
xmlHttpRequest.onreadystatechange = () => { };
xmlHttpRequest.send();

Lostark Open API POST Request Example
POST /auction/items API requires a 'requestAuctionItems' object as a POST body. You can specify an object body with -d in curl.


POST Body curl example
curl -X 'POST' 'https://developer-lostark.game.onstove.com/markets/items'
  -H 'accept: application/json'
  -H 'authorization: bearer your_JWT'
  -H 'Content-Type: application/json'
  -d '{
  "CategoryCode": 20000
}'
POST Body Javascript example
var xmlHttpRequest = new XMLHttpRequest();
xmlHttpRequest.open("POST", "https://developer-lostark.game.onstove.com/markets/items", true);
xmlHttpRequest.setRequestHeader('accept', 'application/json');
xmlHttpRequest.setRequestHeader('authorization', 'bearer your_JWT');
xmlHttpRequest.setRequestHeader('content-Type', 'application/json');
xmlHttpRequest.onreadystatechange = () => { };
xmlHttpRequest.send(JSON.stringify({"CategoryCode" : 20000}));
API Errors
The Lostark Open API returns basic HTTP errors. See this table below.

Code	Description
200	
OK

401	
Unauthorized

403	
Forbidden

404	
Not Found

415	
Unsupported Media Type

429	
Rate Limit Exceeded

500	
Internal Server Error

502	
Bad Gateway

503	
Service Unavailable

504	
Gateway Timeout

API Requests during Maintenance Hours
During maintenance, the Maintenance page will be displayed across the website, and your requests will return a 503 Service Unavailable HTTP status code. It is advisable to refrain from further requests or adjust the request interval upon encountering this status code.

Inefficient API Requests
Ensure your request quota is utilized effectively! Due to throttling, it's essential for your application to implement a robust caching strategy. Efficient management of in-game data is crucial; for certain datasets, you may only need to call the API once a day or even once a week. For instance:

GET /news/events
GET /auctions/options
GET /markets/options
The response data for these APIs typically remains static unless unexpected maintenance that alters the current resources or response model. Exceeding your quota by continuously generating requests will result in application throttling, necessitating a wait for quota refresh. It is advisable to refrain from excessive polling with microsecond intervals or redundant requests triggered by each user interaction.

API Request Limit
We facilitate client request rate management by providing access to counters and timestamps. These response headers indicate the number of requests allowed per minute, the remaining requests, and the time of the next quota refresh:


Reference reponse header
Response header key	Response header value	Desc
X-RateLimit-Limit	100	Request per minute
X-RateLimit-Remaining	15	The number of available requests for your client.
X-RateLimit-Reset	1668659557	The UNIX timestamp for the next refresh of your quota.
API Status
We have 5 api status.
 means the server is trying to get the status metadata.
 means the server is not operational.
 means the server is partically functional.
 means the server is fully operational.
 means the server is under maintenance.

API Versioning
Be sure to check changelog often to stay updated.

We increment versions when we:

Add new endpoints
Remove endpoints
Make incompatible API changes

Deprecation
API endpoints, properties/fields in the response data, and parameters can all be deprecated for various reasons. Unfortunately, we may not always inform you in advance due to sudden changes in game data.

# CHARACTERS
Schemes

https
Authorize
Characters


GET
/characters/{characterName}/siblings
Returns all character profiles for an account.

# ARMORIES
Schemes

https
Authorize
Armories


GET
/armories/characters/{characterName}
Returns a summary of profile information by a character name.


GET
/armories/characters/{characterName}/profiles
Returns a summary of the basic stats by a character name.


GET
/armories/characters/{characterName}/equipment
Returns a summary of the items equipped by a character name.


GET
/armories/characters/{characterName}/avatars
Returns a summary of the avatars equipped by a character name.


GET
/armories/characters/{characterName}/combat-skills
Returns a summary of the combat skills by a character name.


GET
/armories/characters/{characterName}/engravings
Returns a summary of the engravings equipped by a character name.


GET
/armories/characters/{characterName}/cards
Returns a summary of the cards equipped by a character name.


GET
/armories/characters/{characterName}/gems
Returns a summary of the gems equipped by a character name.


GET
/armories/characters/{characterName}/colosseums
Returns a summary of the proving grounds by a character name.


GET
/armories/characters/{characterName}/collectibles
Returns a summary of the collectibles by a character name.


GET
/armories/characters/{characterName}/arkpassive
Returns a summary of the ark passive by a character name.


GET
/armories/characters/{characterName}/arkgrid
Returns a summary of the ark grid by a character name.


Models
ArmoryProfile
Expand Operations
Stat
Expand Operations
Tendency
Expand Operations
Decoration
Expand Operations
ArmoryEquipment
Expand Operations
ArmoryAvatar
Expand Operations
ArmorySkill
Expand Operations
SkillTripod
Expand Operations
SkillRune
Expand Operations
ArmoryEngraving
Expand Operations
Engraving
Expand Operations
EngravingEffect
Expand Operations
ArkPassiveEffect
Expand Operations
ArmoryCard
Expand Operations
Card
Expand Operations
CardEffect
Expand Operations
Effect
Expand Operations
ArmoryGem
Expand Operations
Gem
Expand Operations
ArmoryGemEffect
Expand Operations
GemEffect
Expand Operations
ColosseumInfo
Expand Operations
Colosseum
Expand Operations
AggregationTeamDeathMatchRank
Expand Operations
AggregationTeamDeathMatch
Expand Operations
AggregationElimination
Expand Operations
Aggregation
Expand Operations
AggregationOneDeathmatch
Expand Operations
Collectible
Expand Operations
CollectiblePoint
Expand Operations
ArkPassive
Expand Operations
ArkPassivePoint
Expand Operations
ArkPassiveEffectSkill
Expand Operations
ArkGrid
Expand Operations
ArkGridSlot
Expand Operations
ArkGridGem
Expand Operations
ArkGridEffect

# AUCTIONS
Schemes

https
Authorize
Auctions


GET
/auctions/options
Returns search options for the auction house.


POST
/auctions/items
Returns all active auctions with search options.


Models
AuctionOption
Expand Operations
SkillOption
Expand Operations
Tripod
Expand Operations
EtcOption
Expand Operations
EtcSub
Expand Operations
EtcValue
Expand Operations
Category
Expand Operations
CategoryItem
Expand Operations
RequestAuctionItems
Expand Operations
SearchDetailOption
Expand Operations
Auction
Expand Operations
AuctionItem
Expand Operations
AuctionInfo
Expand Operations
ItemOption

# MARKETS
Schemes

https
Authorize
Markets


GET
/markets/options
Returns search options for the market.


GET
/markets/items/{itemId}
Returns a market item by ID.


POST
/markets/items
Returns all active market items with search options.


POST
/markets/trades
Returns recently traded market items with search options.


Models
MarketOption
Expand Operations
Category
Expand Operations
CategoryItem
Expand Operations
MarketItemStats
Expand Operations
MarketStatsInfo
Expand Operations
RequestMarketItems
Expand Operations
MarketList
Expand Operations
MarketItem
Expand Operations
MarketTradeList
Expand Operations
TradeMarketItem

# GAMECONTENTS
Schemes

https
Authorize
GameContents


GET
/gamecontents/calendar
Returns a list of Calendar this week.


Models
ContentsCalendar
Expand Operations
LevelRewardItems
Expand Operations
RewardItem
