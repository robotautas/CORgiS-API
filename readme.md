Written to communicate with particular controler (Arduino Nano).
The board accepts command <GET_ALL;> and returns line like "V00=0;V01=0;V02=0; ... S01=00;PUMP=0;" with values from sensors, which have to be stored in time series database. Data collecting loop starts concurently, constantly writing values to databases. Meanwhile, simple API runs on the main routine. It accepts HTTP GET request with parameters, and sends a <SET_...> command to the board to change value accordingly. Request comes back empty due to unpredictable nature of boards stdout, TO BE CONTINUED :)

