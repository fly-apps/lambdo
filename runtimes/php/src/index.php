<?php

try {
    $eventString = file_get_contents(getenv('EVENTS_PATH'));
    if ($eventString === false) {
        throw new \Exception("Could not load events from: ".getenv('EVENTS_PATH'));
    }
    $events = json_decode(json: $eventString, associative: true, flags: JSON_THROW_ON_ERROR);

    $handler = require_once("/app/index.php");

    if (! is_callable($handler)) {
        throw new \Exception("No valid handler found, got: ".gettype($handler));
    }
    
    foreach($events as $event) {
        try {
            $handler($event);
        } catch(\Exception $e) {
            echo "handler execution error: " . $e->getMessage();
        }
    }

} catch(\Exception $e) {
    echo "error: " . $e->getMessage();
    exit(1);
}