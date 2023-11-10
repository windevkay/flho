# FLHO 🚀

## WHY USE FLHO

FLHO is a light-weight state machine built with GO to take advantage of the concurrency features presented by GO-routines. FLHO can help with the following:

- **Reduce the amount of message/topics you need to have in a typical SOA**. Services can inquire about less-crucial state changes on demand, without having to be setup as a 'consumer' for it.
- **Business-critical operations can have workflows that are 'time-aware'**. Certain types of failures (e.g Network) could prematurely terminate an operation. Having an external service such as FLHO is akin to having a 'friendly neighbor' that takes an action if they dont see/hear from you after a while. This vastly eliminates the need for you to implement solutions around retry logic, such as outsourcing to a scheduled job or queue.

## HOW FLHO WORKS

1. _Create a workflow for a specific entity or resource within your application_

This can be for a user entity, shopping cart etc. In the create payload, provide the following information:

- `name (str)`: What you would like to call this workflow
- `states (array)`: What states are you interested in sharing or tracking e.g a shopping cart could be in a state of `HAS_ITEMS_BUT_ABANDONED`, `HAS_ITEMS_ON_PROMOTION`'
- `startingState (str)`: Default starting state (must be in states array)
- `endingState (str)`: Default ending state (must be in states array)
- `isTimed (bool)`: Determines if this workflow is time bound or not
- `timeout (int) | optional`: Acceptable amount of time (hrs or mins) after which if there is no call/trigger to update the state of a run for this workflow, a provided webhook or endpoint should be called
- `timeoutUnit (str: HRS | MIN) | optional`: HRS | MIn
- `webhook (str) | optional`: URL to alert

2. _Trigger a workflow (using an identifier) and commence a 'run' of that workflow_

When triggering the workflow, provide data in a context field within the payload and add as much or as little information as you feel is required for that run. FLHO even allows you to specify the state you would like the run to commence in, else it uses the pre-configured `startingState`.

3. _End the run of a workflow anytime_

Ideally after an operation has completed successfully within your application. In the event its a timed run, if no trigger to end or change state is received after the specified interval, FLHO will post to the provided webhook with the payload that was provied at time of triggering the run.
