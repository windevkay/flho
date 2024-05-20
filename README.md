# FLHO 🚀

FLHO (pronounced as _flow_) is a state machine and workflow engine that can bring efficiencies to decoupled architectures, boost reliability of critical features within an app and provide metrics around bottlenecks within a process or workflow.

## USE CASES
- ### REDUCING THE NEED FOR TOPICS
  A workflow can have multiple transitory states between when it begins and eventually completes. In a decoupled architecture, where multiple services might be interested in those states, there will typically be a need to create _topics_ which services can subscribe to in order to listen for changes. Depending on how many of such workflows are available in an application, this can quickly lead to the creation of numerous topics and a subsequent problem of how to effectively manage said topics. On the part of the subscribed services, there might also be a need to trigger an action, dependent on another internal process running within the service. Meaning that topics published before the process is ready, end up being useless or need to be requeried somehow.

  _FLHO_ solves the above scenario(s) by providing a _runID_ once a workflow begins. Hence, only one topic is required (to share the runID) as opposed to creating and maintaining one topic per state. All interested services can then query for the state of the workflow on demand and when they deem it useful*.

- ### SCHEDULED RETRIES FOR FLAKY 3RD PARTY APIs
  An application can have a number of mission critical functions which might also influence the SLA of the business. Things can get tricky when some of these functions have a 3rd party dependency (e.g. API). Most apps will have some level of retry mechanism built in when working with 3rd party APIs. But its not always enough. If an exponential backoff strategy and/or circuit breaker patten proves ineffective, there might be a need to simply place the retry on a much later schedule.

  _FLHO_ helps with this by allowing users to define retry variables when they create workflows. This includes a retry URL and duration after which to do a retry. When creating the workflow, you can specify an error state (e.g. ENDPOINT_FAILURE) and then within the error handling logic of your code, you can notify FLHO of this state, which would then make the other retry variables kick in.

- ### METRICS ON WORKFLOW BOTTLENECKS
  For long running workflows, it can sometimes be difficult to determine where things tend to slow down or points where there is a bottleneck. Factors that cause this can range from a service handling a lot of load (e.g. not scaling properly) or it can be human factors in which an action is needed in order for things to move forward.

  When working with _FLHO_, you can get some insight into this from a handy metrics dashboard that aggregates and analyzes the runs of that workflow over a given period of time. Providing you with information that you can then take appropriate action on.
