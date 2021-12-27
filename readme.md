# Summary

This is an API server for a user interface to interact with a board, controlling sample preparation system for radiocarbon dating.

- It constantly streams values of all board parameters and writes them to time series database. Later the data will be used for live visualisation on the interface program (done).

- It accepts single commands to change parameters on a board (done).

- It is able to accept sets of instructions and excecute them concurrently, spawning self managed routines (in progress).
