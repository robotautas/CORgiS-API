Originally written to communicate with particular controler (Arduino Nano), but remained unused. So i kept it for further reference :)
The board accepted command <GET_ALL;> and returned line like "V00=0;V01=0;V02=0; ... S01=00;PUMP=0;" with values from sensors, which had to be stored in time series database
(influxdb v1). Besides this main loop, program has some additional functions, like recognize the device by its serial number, check if database is present, create one if not. 