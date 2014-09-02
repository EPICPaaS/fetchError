-- SELECT _ascii 0x1F;

LOAD DATA LOCAL INFILE './out' INTO TABLE logcollection 
    CHARACTER SET UTF8 
    FIELDS TERMINATED BY X'1F'
    LINES TERMINATED BY X'1E'
    (LogTime, ErrorClass, ErrorLine, LogType, IP, Port,LogModule,LogAppID,Stack,LogFilePath);