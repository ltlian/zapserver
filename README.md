# Multithreaded logger for TV viewership statistics

This was an assignment for the University of Stavanger where we were to implement a client that listened for and parsed TV channel change events. These events were a recorded dataset which were played back by a server which no longer exists; this project is for reference only.

The project was carried out in collaboration with Øystein Langeland Sandvik. We were given a scaffold codebase as a starting point.

Some error handling and thread functionality is implemented, but these are areas I would expand upon given more time. The mutex locking is particularily simple and can be expanded to increase thread safety and decrease blocking.
