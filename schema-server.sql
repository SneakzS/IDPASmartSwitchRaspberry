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
    wireID INTEGER NOT NULL,
    sampleTime TIMESTAMP NOT NULL,
    workloadW INTEGER NOT NULL,
    PRIMARY KEY(wireID, sampleTime),
    FOREIGN KEY (wireID) REFERENCES Wire(wireID)
);