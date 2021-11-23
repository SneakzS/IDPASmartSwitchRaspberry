CREATE TABLE WorkloadPlan (
    workloadPlanID INTEGER PRIMARY KEY NOT NULL,
    workloadW INTEGER NOT NULL,
    durationM INTEGER NOT NULL,
    toleranceDurationM INTEGER NOT NULL,
    monthFlags INTEGER NOT NULL,
    dayFlags INTEGER NOT NULL,
    hourFlags INTEGER NOT NULL,
    minuteFlags INTEGER NOT NULL,
    weekdayFlags INTEGER NOT NULL,
    isEnabled BOOLEAN NOT NULL,
);

CREATE TABLE Workload (
    workloadID INTEGER PRIMARY KEY NOT NULL,
    workloadPlanID INTEGER NOT NULL,
    workloadW INTEGER NOT NULL,
    startTime TIMESTAMP NULL,
    endTime TIMESTAMP NULL,
    FOREIGN KEY (workloadPlanID) REFERENCES WorkloadPlan(workloadPlanID)
);