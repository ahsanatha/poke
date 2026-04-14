# KataML Reference for `bot.yml`

This repository treats `bot.yml` as the single source of truth for the KataML bot configuration. The file already follows the schema expected by Kata Platform: metadata at the top, a `flows` block, `nlus`, and a `config` placeholder. This document explains each section with concrete references to how `bot.yml` is structured today.

## 1. Root schema
- `name`, `tag`, `desc`, `timezone`, `deleted_at`: deployment metadata; keep them synced with the Kata dashboard.
- `flows`: contains `greetings`, `pokeInfo`, and `fallback`. Each flow nests `intents`, `states`, `actions`, and optional `methods` (currently empty).
- `nlus`: keyword NLUs matching welcome text, yes/no replies, and Pokemon intents.
- `config`: unused but left defined for future configuration.

## 2. `greetings` flow (user onboarding)

1. **Intents**
   - `startTelegram`: triggered when the user sends `/start` (`condition: content == "/start"`).
   - `intial_greet`: matches greetings via the `greeting` keyword NLU (`match: hello`).
   - `confirmIntent`/`denyIntent`: use the `yesno` keyword NLU to map “yes/iyaa” etc.

2. **States and transitions**
   - `init` immediately sends the user to `askName` with `fallback: true`.
   - `askName` stores `context.name` from `content` and runs the `askName` action (text prompt). The transition to `gotName` only fires when `content != "/start"`. This prevents `/start` from being treated as a name.
   - `gotName` confirms via `confirmName`, which runs `saveUserInfo` and `closeGreeting` actions before ending the flow.

3. **Actions**
   - Text replies (`askName`, `confirmName`, `closeGreeting`, `reAskName`) follow the example conversation from Kata’s documentation, including the personalized welcome message.
   - `saveUserInfo` (type: `api`) posts to `https://hc4k4ccck4kwskc00gw44ggk.triki.cloud/api/v1/register` so the Go backend stores the name.

## 3. `pokeInfo` flow (information request)

1. **Intent**
   - `pokeInfo` uses the `askPokemon` keyword NLU (`match: pokemon`) to catch “Pokemon information” or similar phrases.

2. **States**
   - `init` transitions to `askPoke`, which always replies "Which Pokemon do you need the info for?".
   - When `context.pokeName` is set from the user reply, `getPokeInfo` runs `callPokeAPI` and branches on the returned flags:
     * `result.success == true`: send the descriptive text (`explainPokemon`) and show the image via `pokeImage`.
     * `result.success == false`: send `pokeLost`, which gracefully says the Pokemon isn’t available.

3. **Actions**
   - `callPokeAPI` posts `{ pokemon: $(context.pokeName) }` to `https://hc4k4ccck4kwskc00gw44ggk.triki.cloud/api/v1/pokemon`, expecting the Go service to reply with `{ success: true, data: {...} }` or `{ success: false, message: "not found" }`.
   - `explainPokemon` and `pokeImage` pull fields from `$(result.data.*)` (name/type/height/weight/image) to recreate the example conversation (text + picture) referenced in Kata’s doc.

## 4. `fallback` flow
- Single `fallbackIntent` defined as `text` with `fallback: true`, so when no other flow handles a message the bot replies "i cannot understand you yet…".

## 5. NLUs
- `yesno`: keyword nlu mapping affirmations/denials.
- `askPokemon`: keyword nlu that recognizes phrases such as "poke info", "information", or the misspelled variations.
- `greeting`: keyword nlu matching greetings like "hi", "hello", "hai".

## 6. Alignment guidance
- Always update `nlus` or `actions` in this document when you change them in `bot.yml`.
- Follow the schema explained in the Kata docs (intents → states → transitions → actions) and keep action names descriptive (`askName`, `callPokeAPI`, etc.).
- Use the text/value templates shown here to match the concrete flows (personalized greetings, Pokemon response text, missing-Pokemon message) whenever you modify the conversation.
