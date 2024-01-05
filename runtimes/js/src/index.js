const fs = require('fs');

try {
    const eventsString = fs.readFileSync(process.env.EVENTS_PATH)
    const events = JSON.parse(eventsString)
    const eventsCount = events.length

    const handler_module = require('/app/index.js');

    for (let key in events) {
        try {
            handler_module.handler(events[key]);
        } catch(e) {
            console.error("handler execution error", e)
        }
    }
} catch (e) {
    console.error("error retrieving or running events", e)
    process.exit(1)
}


