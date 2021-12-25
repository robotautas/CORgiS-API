# Summary

This is an API server for other software to interact with a board, controlling sample preparation system for radiocarbon dating. Developing is still in progress.

- It constantly streams values of all board parameters and writes them to time series database. Later the data will be used for live visualisation on the interface program (done).

- It accepts single commands to change parameters on a board (done).

- It is able to spawn sets of instructions concurently as a self managed routine (in progress).
