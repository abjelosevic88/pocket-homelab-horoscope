# API

JSON over HTTP under `/api/v1`. Authenticate with the session cookie (browser), `Authorization: Bearer <token>`
(`phr_…`, made under Settings → API & shortcuts), or a Pocket Core token. Types: `web/src/lib/types.ts`.
Everything is per person: another person's ids answer 404. Times of sky events are RFC 3339 in the person's home zone
(Settings → Charts → Where you live, else their own birth place).

## Your day

| Call | What |
| :--- | :--- |
| `GET /today?date=&person=` | The day's reading for you (or `person`): headline, summary, scores (love, work, mood, energy, overall), the Moon (sign, phase, house, void of course), the transits that count, do / avoid, retrogrades, the day's sky events, the next 7 days, a kept AI reading and your check-in. `{needs_me: true}` until you add yourself |
| `GET /transits?person=&from=&days=&slow=1` | Transits of the coming days (≤ 730): start, exact passes (up to three with retrogrades), end, text, tone |
| `GET /dashboard` | Short overview (Moon, headline, scores, next sky events, your Big Three) |
| `GET /day/{date}` | One day for Pocket Day: the Moon, headline, scores, events, your check-in |
| `GET /overlay?from=&to=` | Calendar items for Pocket Calendar: lunations, eclipses, stations, ingresses, and the exact days of slow transits to you |

## People and charts

| Call | What |
| :--- | :--- |
| `GET /people?q=` · `POST /people` · `GET/PUT/DELETE /people/{id}` | People with their Big Three, age and next birthday. Body: `name relation is_me birth_date birth_time place country lat lon tz color notes favorite` (`tz`: IANA name or `UTC+01:00`) |
| `GET /people/{id}/chart` | The birth chart: positions, houses, aspects, element balance, and the text for each placement |
| `POST /people/preview` | The Big Three of unsaved birth data |
| `POST /people/import` `{text}` | AAF (`#A93/#B93`, astro.com, Solar Fire) or CSV (`name,date,time,place,lat,lon,tz,relation`) → `{added, skipped, problems}` |
| `GET /people/contacts` · `POST /people/contacts` `{ids, relation}` | Your contacts with a birthday in Pocket Contacts; add them as people |
| `GET /places?q=` | Place search (name, region, country, lat, lon, tz) |
| `GET /match?a=&b=` | Compatibility of two people (a = you by default): scores, strengths, challenges, every aspect |

## The sky

| Call | What |
| :--- | :--- |
| `GET /sky/now?at=` | Where everything is now, the Moon, a chart for your place, the next two weeks' events |
| `GET /sky/events?from=&to=&kinds=phase,eclipse,ingress,station&moon=1` | Sky events (≤ 2 years) |
| `GET /sky/moon?month=2026-10` | The Moon calendar: phase, sign and sign change, events, void-of-course and your mood per day |
| `GET /sky/retrogrades?year=` | Retrograde periods of a year |

## Journal and readings

| Call | What |
| :--- | :--- |
| `GET /journal?from=&to=&limit=` · `POST /journal` `{date, mood, energy, note, tags}` · `PUT/DELETE /journal/{id}` | One check-in a day (a second one replaces it); the day's sky is kept with it |
| `GET /journal/insights?days=` | Mood and energy by Moon phase, element and sign; with Pocket Health linked, sleep and resting heart rate by phase |
| `POST /readings` `{kind: daily\|natal\|match\|year\|question, person_id, other_id, date, question, fresh}` | A written reading from your AI (Pocket AI or OpenAI) → `{ok, reading}` / `{ok: false, not_configured}` |
| `GET /readings?kind=&person=&saved=1` · `GET/PUT/DELETE /readings/{id}` | Readings; `PUT {saved}` keeps one (unsaved ones go after 60 days) |
| `POST /readings/{id}/note` | Keep a reading as a note in Pocket Notes |
| `POST /ai/ask` `{question, person_id}` · `GET /ai/status` · `GET /ai/models` | |

## Account and box

`GET /auth/status`, `POST /auth/login`, `POST /auth/setup`, `POST /auth/logout`, `GET|PUT /auth/me`, `GET|POST /tokens`,
`DELETE /tokens/{id}`, `GET|PUT /settings`, `GET /settings/system`, `POST /integrations/test`, `POST /notify/test`,
`POST /notify/today`, `GET|POST /users`, `PUT|DELETE /users/{id}` (admins), `GET /backup` (admins), `GET /export.json`,
`GET /export.csv` (people), `POST /import` (a Pocket Horoscope export).

`GET /healthz`, `GET /readyz`, `GET /metrics`.

## Pocket Core links

Offers `horoscope:read` (`/dashboard /overlay /today /transits /sky/now /sky/events /sky/moon /journal /day/{date}`) and
`horoscope:write` (`POST /journal`, `PUT/DELETE /journal/{id}` — only check-ins that app wrote). Publishes
`horoscope.daily_ready`, `horoscope.journal_logged`, `horoscope.person_added`.
Uses Pocket Contacts (`contacts:read`), Pocket Notes (`notes:write`), Pocket Health (`health:read`) and Pocket AI.
