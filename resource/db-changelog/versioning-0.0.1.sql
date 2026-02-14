create table `users`
(
    `id`          bigint unsigned NOT NULL AUTO_INCREMENT,
    `code`        varchar(50),
    `profile`     varchar(255),
    `full_name`   varchar(150),
    `phone`       varchar(15),
    `email`       varchar(60) NOT NULL,
    `password`    varchar(255),
    `is_verified` tinyint(1) DEFAULT '0',
    `otp_code`    int default null,
    `auth_code`   varchar(200) null,
    `lang`        varchar(10) null,
    `last_active` timestamp null,
    `created_at`  timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `created_by`  varchar(150),
    `updated_by`  varchar(150),
    primary key (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2033 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Summary Homepage Table
CREATE TABLE IF NOT EXISTS summary_homepage
(
    id
    BIGINT
    UNSIGNED
    AUTO_INCREMENT
    PRIMARY
    KEY,
    terapis
    INT
    NOT
    NULL
    DEFAULT
    0,
    kota
    INT
    NOT
    NULL
    DEFAULT
    0,
    rating
    INT
    NOT
    NULL
    DEFAULT
    0,
    created_at
    TIMESTAMP
    DEFAULT
    CURRENT_TIMESTAMP,
    updated_at
    TIMESTAMP
    DEFAULT
    CURRENT_TIMESTAMP
    ON
    UPDATE
    CURRENT_TIMESTAMP,
    created_by
    VARCHAR
(
    100
),
    updated_by VARCHAR
(
    100
)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE =utf8mb4_unicode_ci;

-- Master Booking Status Table
CREATE TABLE IF NOT EXISTS master_booking_statuses
(
    id
    BIGINT
    UNSIGNED
    AUTO_INCREMENT
    PRIMARY
    KEY,
    name
    VARCHAR
(
    50
) NOT NULL,
    code VARCHAR
(
    50
) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by VARCHAR
(
    100
),
    updated_by VARCHAR
(
    100
),
    INDEX idx_code
(
    code
)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE =utf8mb4_unicode_ci;

-- Insert default booking statuses
INSERT INTO master_booking_statuses (name, code, created_at, updated_at)
VALUES ('Pending', 'pending', NOW(), NOW()),
       ('Ongoing', 'ongoing', NOW(), NOW()),
       ('Completed', 'completed', NOW(), NOW()),
       ('Canceled', 'canceled', NOW(), NOW()) ON DUPLICATE KEY
UPDATE name=
VALUES (name);

-- Bookings Table
CREATE TABLE IF NOT EXISTS bookings
(
    id
    BIGINT
    UNSIGNED
    AUTO_INCREMENT
    PRIMARY
    KEY,
    user_id
    BIGINT
    UNSIGNED
    NOT
    NULL,
    therapist_id
    BIGINT
    UNSIGNED
    NOT
    NULL,
    city_id
    BIGINT
    UNSIGNED
    NOT
    NULL,
    status_id
    BIGINT
    UNSIGNED
    NOT
    NULL,
    date_time
    TIMESTAMP
    NOT
    NULL,
    created_at
    TIMESTAMP
    DEFAULT
    CURRENT_TIMESTAMP,
    updated_at
    TIMESTAMP
    DEFAULT
    CURRENT_TIMESTAMP
    ON
    UPDATE
    CURRENT_TIMESTAMP,
    created_by
    VARCHAR
(
    100
),
    updated_by VARCHAR
(
    100
),
    INDEX idx_user_id
(
    user_id
),
    INDEX idx_therapist_id
(
    therapist_id
),
    INDEX idx_city_id
(
    city_id
),
    INDEX idx_status_id
(
    status_id
),
    INDEX idx_date_time
(
    date_time
),
    FOREIGN KEY
(
    user_id
) REFERENCES users
(
    id
) ON DELETE CASCADE,
    FOREIGN KEY
(
    therapist_id
) REFERENCES therapists
(
    id
)
  ON DELETE CASCADE,
    FOREIGN KEY
(
    city_id
) REFERENCES master_cities
(
    id
)
  ON DELETE CASCADE,
    FOREIGN KEY
(
    status_id
) REFERENCES master_booking_statuses
(
    id
)
  ON DELETE RESTRICT
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE =utf8mb4_unicode_ci;

-- Booking Status History Table
CREATE TABLE IF NOT EXISTS booking_status_histories
(
    id
    BIGINT
    UNSIGNED
    AUTO_INCREMENT
    PRIMARY
    KEY,
    booking_id
    BIGINT
    UNSIGNED
    NOT
    NULL,
    status_id
    BIGINT
    UNSIGNED
    NOT
    NULL,
    notes
    TEXT,
    created_at
    TIMESTAMP
    DEFAULT
    CURRENT_TIMESTAMP,
    updated_at
    TIMESTAMP
    DEFAULT
    CURRENT_TIMESTAMP
    ON
    UPDATE
    CURRENT_TIMESTAMP,
    created_by
    VARCHAR
(
    100
),
    updated_by VARCHAR
(
    100
),
    INDEX idx_booking_id
(
    booking_id
),
    INDEX idx_status_id
(
    status_id
),
    FOREIGN KEY
(
    booking_id
) REFERENCES bookings
(
    id
) ON DELETE CASCADE,
    FOREIGN KEY
(
    status_id
) REFERENCES master_booking_statuses
(
    id
)
  ON DELETE RESTRICT
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE =utf8mb4_unicode_ci;

-- Settings Table
CREATE TABLE IF NOT EXISTS settings
(
    id
    BIGINT
    UNSIGNED
    AUTO_INCREMENT
    PRIMARY
    KEY,
    application_fee
    DECIMAL
(
    10,
    2
) NOT NULL DEFAULT 0.00,
    tax_ppn DECIMAL
(
    5,
    2
) NOT NULL DEFAULT 0.00,
    service_fee DECIMAL
(
    10,
    2
) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by VARCHAR
(
    100
),
    updated_by VARCHAR
(
    100
)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE =utf8mb4_unicode_ci;

-- Insert default settings
INSERT INTO settings (application_fee, tax_ppn, service_fee, created_at, updated_at)
VALUES (5000.00, 11.00, 2500.00, NOW(), NOW()) ON DUPLICATE KEY
UPDATE application_fee=
VALUES (application_fee);

-- Payments Table
CREATE TABLE IF NOT EXISTS payments
(
    id
    BIGINT
    UNSIGNED
    AUTO_INCREMENT
    PRIMARY
    KEY,
    total
    DECIMAL
(
    12,
    2
) NOT NULL DEFAULT 0.00,
    reference_number VARCHAR
(
    100
) NOT NULL UNIQUE,
    booking_id BIGINT UNSIGNED NOT NULL,
    method VARCHAR
(
    50
) NOT NULL,
    status VARCHAR
(
    50
) NOT NULL,
    third_party_id BIGINT UNSIGNED,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by VARCHAR
(
    100
),
    updated_by VARCHAR
(
    100
),
    INDEX idx_booking_id
(
    booking_id
),
    INDEX idx_reference_number
(
    reference_number
),
    INDEX idx_status
(
    status
),
    FOREIGN KEY
(
    booking_id
) REFERENCES bookings
(
    id
)
                                                   ON DELETE CASCADE
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE =utf8mb4_unicode_ci;

-- Payment Details Table
CREATE TABLE IF NOT EXISTS payment_details
(
    id
    BIGINT
    UNSIGNED
    AUTO_INCREMENT
    PRIMARY
    KEY,
    amount
    DECIMAL
(
    12,
    2
) NOT NULL DEFAULT 0.00,
    reference_number VARCHAR
(
    100
) NOT NULL,
    name VARCHAR
(
    100
) NOT NULL,
    parent_payment_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by VARCHAR
(
    100
),
    updated_by VARCHAR
(
    100
),
    INDEX idx_parent_payment_id
(
    parent_payment_id
),
    INDEX idx_reference_number
(
    reference_number
),
    FOREIGN KEY
(
    parent_payment_id
) REFERENCES payments
(
    id
)
                                                   ON DELETE CASCADE
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE =utf8mb4_unicode_ci;

-- Master Service Categories Table
CREATE TABLE IF NOT EXISTS master_service_categories
(
    id
    BIGINT
    UNSIGNED
    AUTO_INCREMENT
    PRIMARY
    KEY,
    name
    VARCHAR
(
    100
) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by VARCHAR
(
    100
),
    updated_by VARCHAR
(
    100
)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE =utf8mb4_unicode_ci;

-- Insert default service categories
INSERT INTO master_service_categories (name, created_by, updated_by, created_at, updated_at)
VALUES ('Rehabilitation', 'system', 'system', NOW(), NOW()),
       ('Massage', 'system', 'system', NOW(), NOW()),
       ('Sports Therapy', 'system', 'system', NOW(), NOW()),
       ('Pain Management', 'system', 'system', NOW(), NOW()) ON DUPLICATE KEY
UPDATE name=
VALUES (name);

-- Services Table
CREATE TABLE IF NOT EXISTS services
(
    id
    BIGINT
    UNSIGNED
    AUTO_INCREMENT
    PRIMARY
    KEY,
    name
    VARCHAR
(
    255
) NOT NULL,
    category_id BIGINT UNSIGNED NOT NULL,
    description TEXT,
    duration INT NOT NULL DEFAULT 60,
    price DECIMAL
(
    15,
    2
) NOT NULL DEFAULT 0.00,
    commission DECIMAL
(
    5,
    2
) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by VARCHAR
(
    100
),
    updated_by VARCHAR
(
    100
),
    INDEX idx_category_id
(
    category_id
),
    INDEX idx_name
(
    name
),
    FOREIGN KEY
(
    category_id
) REFERENCES master_service_categories
(
    id
)
                                                   ON DELETE RESTRICT
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE =utf8mb4_unicode_ci;

-- Service Areas Table (Junction table for Service-City many-to-many)
CREATE TABLE IF NOT EXISTS service_areas
(
    id
    BIGINT
    UNSIGNED
    AUTO_INCREMENT
    PRIMARY
    KEY,
    service_id
    BIGINT
    UNSIGNED
    NOT
    NULL,
    city_id
    BIGINT
    UNSIGNED
    NOT
    NULL,
    created_at
    TIMESTAMP
    DEFAULT
    CURRENT_TIMESTAMP,
    updated_at
    TIMESTAMP
    DEFAULT
    CURRENT_TIMESTAMP
    ON
    UPDATE
    CURRENT_TIMESTAMP,
    created_by
    VARCHAR
(
    100
),
    updated_by VARCHAR
(
    100
),
    INDEX idx_service_id
(
    service_id
),
    INDEX idx_city_id
(
    city_id
),
    UNIQUE KEY unique_service_city
(
    service_id,
    city_id
),
    FOREIGN KEY
(
    service_id
) REFERENCES services
(
    id
) ON DELETE CASCADE,
    FOREIGN KEY
(
    city_id
) REFERENCES master_cities
(
    id
)
  ON DELETE CASCADE
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE =utf8mb4_unicode_ci;

-- Service Included Items Table
CREATE TABLE IF NOT EXISTS service_included_items
(
    id
    BIGINT
    UNSIGNED
    AUTO_INCREMENT
    PRIMARY
    KEY,
    service_id
    BIGINT
    UNSIGNED
    NOT
    NULL,
    name
    VARCHAR
(
    255
) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by VARCHAR
(
    100
),
    updated_by VARCHAR
(
    100
),
    INDEX idx_service_id
(
    service_id
),
    FOREIGN KEY
(
    service_id
) REFERENCES services
(
    id
)
                                                   ON DELETE CASCADE
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE =utf8mb4_unicode_ci;
