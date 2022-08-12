Write a program that can be run by the technicians on the production line just before assembling
the sensor units. They will note the output and feed it into the production system. The
technicians can sometimes be impatient and may kill the process if it takes too long.
A stretch goal (as a future improvement) is an API that the production system can integrate with
directly.
    1. Write an application (CLI) that creates a batch of 100 unique DevEUIs and registers
        them with the LoRaWAN api.
    2. (Stretch) Write an API with an HTTP interface that creates the batch and returns it to the
        client.

Requirements
1) CLI

    a)  The application must return every DevEUI that it registers with the LoRaWAN
        provider (e.g. if the application is killed it must wait for in-flight requests to finish
        otherwise we would have registered those DevEUIs but would not be using them)
        
    b)  It must handle user interrupts gracefully (SIGINT)
    
    c)  It must register exactly 100 DevEUIs (no more) with the provider (to avoid paying
        for DevEUIs that we do not use)
        
    d)  It should make multiple requests concurrently (but there must never be more than
        10 requests in-flight, to avoid throttling)

2) API

    a) The request must be idempotent. It is possible that the production system could
        timeout or make multiple requests simultaneously and each request should return
        then same set of DevEUIs
    b) The response should be a json body with an array of 100 elements {“deveuis”:
        [“FFA45722AA738240”,....]}

LoRaWAN API
The registration operation is an HTTP request to the LoRaWAN provider’s API. The details are
below
host: europe-west1-machinemax-dev-d524.cloudfunctions.net
paths:
/sensor-onboarding-sample:
post:
consumes:
- application/json
parameters:
- in: body
- required: true
- type: string
- name: deveui
responses:
200:
description: The device has been successfully registered
422:
description: The devEUI has already been used
