CREATE TABLE Customer (
    customerID INTEGER NOT NULL PRIMARY KEY,
    firstName TEXT NOT NULL,
    lastName TEXT NOT NULL,
    address TEXT NOT NULL,
    town TEXT NOT NULL
);

CREATE TABLE Wire (
    wireID INTEGER NOT NULL PRIMARY KEY,
    capacityW INTEGER
);

CREATE TABLE WireCustomer (
    wireID INTEGER NOT NULL,
    customerID INTEGER NOT NULL,
    PRIMARY KEY (wireID, customerID),
    FOREIGN KEY (wireID) REFERENCES Wire(wireID),
    FOREIGN KEY (customerID) REFERENCES Customer(customerID)
);

CREATE TABLE WireWorkload (
    wireWorkloadID INTEGER NOT NULL PRIMARY KEY,
    wireID INTEGER NOT NULL,
    workloadW INTEGER NOT NULL,
    startTime TIMESTAMP NOT NULL,
    endTime TIMESTAMP NOT NULL,
    FOREIGN KEY (wireID) REFERENCES Wire(wireID)
);