- [x] design doc
- [x] set up our stack
- [x] base schema
- [x] create room feature
- [x] users join room feature(web socket)
- [x] static quiz for v1

- [x] when a person leaves the room or got disconnected we need to update room member table

 ## starting quiz 
- [x] all users( including bot) joins the room.
- [x] each question duration should be given while room creation

- [x] all users should click start game (bot not needed) room_members table should have a new col status

- [x] first question comes to all
- [x] once solved he can click next and live leaderboard gets updated

- [x] if all submits before the deadline (other than bots) we move to next question or if the question time gets over
- [x] finally we show whats right and whats wrong answer after the end of quiz
- [x] start the bots thingy when all other players clicked start and just before loading the first question

## todo
- [x] go through 1 full flow with static data, 
  - [x] room creation
  - [x] starting game
  - [x] single player flow
  - [x] going to all the questions until end answer
- [x] build the leader board feature
- [x] start with ui

~~- [ ] exiting in the middle of the game and rejoining~~

~~- [ ] when a person leaves the room or got disconnected we need to update room member table~~

- [x] room members room_id column should be renamed to room_code
- [x] fix room members 
- [x] get the bot times drop down from backend and fill the dropdown

- [x] rename answer id to answer option
- [x] pong not working after some time!
- [x] store data in db and fetch
- [x] leaderboard
- [x] free up memory from the map after things are done both user and bots
that even bot can use the same client whenever the want to send
- [x] display the answers user selected when they click analyze button
- [x] points system needs to change, every other attempts you need to reduce points
- [x] fix leaderboard
- [x] find a way to store all inmemory data to pgsql
- [x] store all the answers in db
- [x] i selected only 1 question but all the questions were visible
- [x] if you update answer the existing answer should be updated not added
- [x] fix final leaderboard
- [x] fix game completion
- [x] fix memory issue (how can we store it in db) and clear inmemory data once used 

- [x] other page displays
- [x] scoring needs to be fixed, if he is trying for the second time we need to do -50
game
- [x] update room meta with the final leader board scores as a score json
- [x] need detailed analytics including time it took to answer each question in analytics page based on the picked answer we need to give final analytics comments 

- [x] add exit game button which will exit you out of the game remove ws
- [x] write a timer which will delete the inmemory cache once in an hour
- [x] find a way on how to check for memory leak
- [x] error rendering on screen pop up
- [x] error rendering on screen pop up
- [x] shouldnt be able to exit without ticking a modal in between 
- [x] close the ws connection when he clicks yes to the modal for game analysis
- [x] auth
- [x] session
- [x] remove egress channel and change it into connection map so 
- [x] if i do hard reload after game is over it again goes to the first question
- [x] authentication and session
- [ ] check for deadlock and race conditions
- [ ] productionize tailwind remove tailwind.config from layout
- [ ] validate and sanitize all models
- [ ] tips to imporve using gpt in analysis
- [ ] there is some bug with updating state of the room fix it check all the logic and make sure we are updating the db
- [ ] disable refresh in game or atleast get a pop up. upon refreshjing he shouldf be kicked out of the ws and sent to home
- [ ] close all open channels
- [ ] confetti needs to come only for the winning user
- [ ] fix multiplayer
- [ ] multiplayer
- [ ] go through all fmt.println statements and remove unwanted stuff
- [ ] use safehtml for all the backend data which we are sending to ui
- [ ] xss protection
- [ ] handle ui errors display appropriately especially inside ws
- [ ] if a user leaves before starting a game we need to remove him( ie if his connection is not active or he leaves explicitly)
- [ ] they shouldnt add a bot which takes more time than room each question time
- [ ] there is another checklist shown with the status of others submitted the question or not
- [ ] go through all app todos
- [ ] go through all ui resize it and see if it needs some fixes like auto scrolling
- [ ] cleanup of roomstatus and gamestatus in the manager (we are already cleaning client in the timer)
- [ ] success message should get green pop up
- [ ] navigation bugs fix (using back button) 2025-05-14T10:46:07.350+0530	error	users/user_service.go:19	Could not initialize databasefailed to connect to PostgreSQL: failed to connect to `user=postgres database=postgres`:
	[::1]:15432 (localhost): server error: FATAL: sorry, too many clients already (SQLSTATE 53300)
	127.0.0.1:15432 (localhost): server error: FATAL: sorry, too many clients already (SQLSTATE 53300)
- [ ] completely gamify the ui
- [ ] auth 0 production needs to come from the official oauth keys https://auth0.com/docs/authenticate/identity-providers/social-identity-providers/devkeys https://community.auth0.com/t/how-to-move-from-development-key-to-production-key-for-tenant/62860/2 
- [ ] add pagination to my quiz page
- [ ] my quiz page should say the multiplayer quiz as expired if it is created half an hour before
- [ ] abandon state if comes out of quiz room
## v2
- [ ] move to redis
- [ ] uploading pdf's to generate question
## prod todo check
- [ ] tailwind
- [ ] get google auth keys and set it up properly
- [ ] secret vault
- [ ] setup proper docker container which can access the secrets
  
auth0 docs:
https://auth0.com/docs/quickstart/webapp/golang/interactive

## security checklist
Here is a **step-by-step checklist** to make your frontend (HTML + Go + HTMX + Gin) code as secure as practically possible against frontend attacks like **XSS**, **CSRF**, **Clickjacking**, and other common web threats. While nothing is ever **100% secure**, following these steps gives you **best-in-class protection**.

---

## ✅ Frontend Security Checklist (HTMX + Go + Gin)

---

### 🔐 1. **Sanitize All Dynamic Data**

| Step | What to Do                                                                                           | Notes                                                                                                  |                             |
| ---- | ---------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------ | --------------------------- |
| 1.1  | Use Go’s `html/template` (not `text/template`)                                                       | Ensures automatic HTML escaping                                                                        |                             |
| 1.2  | Never use \`{{ .                                                                                     | safeHTML }}\` unless using a sanitizer (e.g. [bluemonday](https://github.com/microcosm-cc/bluemonday)) | Prevents raw HTML injection |
| 1.3  | For JSON in attributes (like `data-hx-headers`), use `template.JS` and manually construct valid JSON | Prevents script context issues                                                                         |                             |

---

### 🛡️ 2. **CSRF Protection (HTMX-Compatible)**

| Step | What to Do                                                                                    | Notes                                        |
| ---- | --------------------------------------------------------------------------------------------- | -------------------------------------------- |
| 2.1  | Generate a per-session CSRF token and store in `HttpOnly`, `Secure`, `SameSite=Strict` cookie | Use `c.SetCookie(...)` in Gin                |
| 2.2  | Echo token in a meta tag or `data-hx-headers` (you are doing this)                            | Use Go's `template.JS` to safely encode JSON |
| 2.3  | Validate CSRF token on each unsafe request (`POST`, `PUT`, etc.) server-side                  | Use middleware or per-route validation       |

---

### 🔒 3. **Set Critical HTTP Security Headers**

Set these in middleware in Gin:

```go
func SecureHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
        c.Header("Content-Security-Policy", "default-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none';")
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
        c.Next()
    }
}
```

Apply it:

```go
r := gin.Default()
r.Use(SecureHeaders())
```

---

### 🧱 4. **Content Security Policy (CSP)**

| Step | What to Do                                         | Notes                                                      |
| ---- | -------------------------------------------------- | ---------------------------------------------------------- |
| 4.1  | Disallow inline scripts unless using a nonce       | Helps mitigate XSS                                         |
| 4.2  | Only allow scripts and styles from trusted sources | Use `'self'` and `nonce-<value>` or external files         |
| 4.3  | Avoid `unsafe-inline`                              | Consider `strict-dynamic` with nonce for Alpine.js or HTMX |

Example (strong CSP):

```http
Content-Security-Policy: default-src 'self'; script-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none';
```

---

### 🧪 5. **Input Validation & Output Escaping**

| Step | What to Do                                             | Notes                                                           |
| ---- | ------------------------------------------------------ | --------------------------------------------------------------- |
| 5.1  | Sanitize or validate all user input on the server side | Use strong schema validation                                    |
| 5.2  | Escape output based on context (HTML, JS, JSON, URLs)  | Go templates help here, but be cautious with embedded JS or CSS |

---

### 🧬 6. **Session and Cookie Hygiene**

| Step | What to Do                                             | Notes                        |
| ---- | ------------------------------------------------------ | ---------------------------- |
| 6.1  | Set cookies as `Secure`, `HttpOnly`, `SameSite=Strict` | Prevents theft via XSS/CSRF  |
| 6.2  | Rotate session IDs after login/logout                  | Prevents session fixation    |
| 6.3  | Store minimal session data client-side                 | Avoid leaking sensitive data |

---

### 🧯 7. **Clickjacking Protection**

Set one of:

```http
X-Frame-Options: DENY
# or, better:
Content-Security-Policy: frame-ancestors 'none';
```

---

### 🔍 8. **Audit Static Files & External Dependencies**

| Step | What to Do                                                     | Notes                                                        |
| ---- | -------------------------------------------------------------- | ------------------------------------------------------------ |
| 8.1  | Serve JS/CSS locally (don’t use CDNs without integrity checks) | Prevent supply-chain attacks                                 |
| 8.2  | Use Subresource Integrity (SRI) if you must load from CDN      | `<script src="..." integrity="..." crossorigin="anonymous">` |
| 8.3  | Periodically audit npm/pnpm packages for vulnerabilities       | `npm audit`, `osv.dev`, etc.                                 |

---

### 🧼 9. **Disable Unnecessary Browser Features**

Use `Permissions-Policy` to lock down:

```http
Permissions-Policy: camera=(), microphone=(), geolocation=(), usb=()
```

---

### 📜 10. **Limit Error Information in HTML**

| Step | What to Do                                          | Notes                      |
| ---- | --------------------------------------------------- | -------------------------- |
| 10.1 | Do not show stack traces or internal errors in HTML | Show user-friendly message |
| 10.2 | Log details server-side                             | Secure logs from tampering |

---

### ✅ Final Sanity Checks

* [ ] HTTPS is enabled everywhere (force-redirect HTTP to HTTPS)
* [ ] Static assets have appropriate cache headers (`Cache-Control`)
* [ ] All user-generated content is validated and sanitized
* [ ] Server-side logs are monitored and stored securely
* [ ] Rate limiting or CAPTCHA is applied to sensitive endpoints

