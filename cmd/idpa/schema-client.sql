CREATE TABLE WorkloadDefinition (
    workloadDefinitionID INTEGER PRIMARY KEY NOT NULL,
    workloadW INTEGER NOT NULL,
    durationM INTEGER NOT NULL,
    toleranceDurationM INTEGER NOT NULL,
    isEnabled BOOLEAN NOT NULL,
    description TEXT NOT NULL,
    expiryDate TEXT NOT NULL
);

CREATE TABLE TimePattern (
    timePatternID INTEGER NOT NULL PRIMARY KEY,
    workloadDefinitionID INTEGER NOT NULL,
    monthFlags INTEGER NOT NULL,
    dayFlags INTEGER NOT NULL,
    hourFlags INTEGER NOT NULL,
    minuteFlags INTEGER NOT NULL,
    weekdayFlags INTEGER NOT NULL,
    FOREIGN KEY (workloadDefinitionID) REFERENCES WorkloadDefinition(workloadDefinitionID) ON DELETE CASCADE
);

CREATE TABLE Workload (
    workloadID INTEGER NOT NULL PRIMARY KEY,
    workloadDefinitionID INTEGER NOT NULL,
    matchTime TIMESTAMP NOT NULL,
    workloadW INTEGER NOT NULL,
    offsetM INTEGER NOT NULL,
    durationM INTEGER NOT NULL,
    FOREIGN KEY (workloadDefinitionID) REFERENCES WorkloadDefinition(workloadDefinitionID),
    UNIQUE (matchTime, workloadDefinitionID)
);

CREATE TABLE WorkloadSample (
    sampleTime TIMESTAMP NOT NULL PRIMARY KEY,
    workloadID INTEGER NOT NULL,
    FOREIGN KEY (workloadID) REFERENCES Workload(workloadID)
);

CREATE TABLE SensorSample (
    sampleTime TIMESTAMP NOT NULL PRIMARY KEY,
    power FLOAT NOT NULL,
    current FLOAT NOT NULL,
    voltage FLOAT NOT NULL,
);

CREATE TABLE Log (
    logID INTEGER NOT NULL PRIMARY KEY,
    logTime TIMESTAMP NOT NULL,
    severity INTEGER NOT NULL,
    source TEXT NOT NULL,
    message TEXT NOT NULL
);