<?php

return function(array $event) {
    echo "Let's do an event: " . json_encode($event) . "\n";
};