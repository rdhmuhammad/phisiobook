SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

CREATE TABLE IF NOT EXISTS `master_roles` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(100) NOT NULL,
  `label` VARCHAR(100) NOT NULL,
  `created_at` DATETIME(3) NULL DEFAULT NULL,
  `updated_at` DATETIME(3) NULL DEFAULT NULL,
  `created_by` VARCHAR(100) NULL DEFAULT NULL,
  `updated_by` VARCHAR(100) NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_master_roles_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `users` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code` VARCHAR(100) NULL DEFAULT NULL,
  `profile` TEXT NULL,
  `full_name` VARCHAR(255) NULL DEFAULT NULL,
  `phone` VARCHAR(50) NULL DEFAULT NULL,
  `email` VARCHAR(255) NOT NULL,
  `password` TEXT NULL,
  `is_verified` TINYINT NOT NULL DEFAULT 0,
  `otp_code` INT NULL DEFAULT NULL,
  `auth_code` VARCHAR(255) NULL DEFAULT NULL,
  `lang` VARCHAR(10) NULL DEFAULT NULL,
  `last_active` DATETIME(3) NULL DEFAULT NULL,
  `chat_ref` VARCHAR(100) NULL DEFAULT NULL,
  `created_at` DATETIME(3) NULL DEFAULT NULL,
  `updated_at` DATETIME(3) NULL DEFAULT NULL,
  `created_by` VARCHAR(100) NULL DEFAULT NULL,
  `updated_by` VARCHAR(100) NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_users_email` (`email`),
  KEY `idx_users_code` (`code`),
  KEY `idx_users_auth_code` (`auth_code`),
  KEY `idx_users_chat_ref` (`chat_ref`),
  KEY `idx_users_is_verified` (`is_verified`),
  KEY `idx_users_last_active` (`last_active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `user_admins` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code` VARCHAR(100) NULL DEFAULT NULL,
  `full_name` VARCHAR(255) NULL DEFAULT NULL,
  `email` VARCHAR(255) NOT NULL,
  `phone` VARCHAR(50) NULL DEFAULT NULL,
  `role_id` BIGINT UNSIGNED NOT NULL,
  `password` TEXT NULL,
  `auth_code` VARCHAR(255) NULL DEFAULT NULL,
  `is_active` TINYINT NOT NULL DEFAULT 1,
  `last_active` DATETIME(3) NULL DEFAULT NULL,
  `chat_ref` VARCHAR(100) NULL DEFAULT NULL,
  `created_at` DATETIME(3) NULL DEFAULT NULL,
  `updated_at` DATETIME(3) NULL DEFAULT NULL,
  `created_by` VARCHAR(100) NULL DEFAULT NULL,
  `updated_by` VARCHAR(100) NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_admins_email` (`email`),
  KEY `idx_user_admins_code` (`code`),
  KEY `idx_user_admins_auth_code` (`auth_code`),
  KEY `idx_user_admins_chat_ref` (`chat_ref`),
  KEY `idx_user_admins_role_id` (`role_id`),
  KEY `idx_user_admins_is_active` (`is_active`),
  KEY `idx_user_admins_last_active` (`last_active`),
  CONSTRAINT `fk_user_admins_role_id` FOREIGN KEY (`role_id`) REFERENCES `master_roles` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `master_cities` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(255) NOT NULL,
  `created_at` DATETIME(3) NULL DEFAULT NULL,
  `updated_at` DATETIME(3) NULL DEFAULT NULL,
  `created_by` VARCHAR(100) NULL DEFAULT NULL,
  `updated_by` VARCHAR(100) NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_master_cities_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `master_booking_statuses` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(50) NOT NULL,
  `code` VARCHAR(50) NOT NULL,
  `created_at` DATETIME(3) NULL DEFAULT NULL,
  `updated_at` DATETIME(3) NULL DEFAULT NULL,
  `created_by` VARCHAR(100) NULL DEFAULT NULL,
  `updated_by` VARCHAR(100) NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_master_booking_statuses_name` (`name`),
  UNIQUE KEY `uk_master_booking_statuses_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `master_therapy_types` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(255) NOT NULL,
  `created_at` DATETIME(3) NULL DEFAULT NULL,
  `updated_at` DATETIME(3) NULL DEFAULT NULL,
  `created_by` VARCHAR(100) NULL DEFAULT NULL,
  `updated_by` VARCHAR(100) NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_master_therapy_types_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `master_service_categories` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(100) NOT NULL,
  `created_at` DATETIME(3) NULL DEFAULT NULL,
  `updated_at` DATETIME(3) NULL DEFAULT NULL,
  `created_by` VARCHAR(100) NULL DEFAULT NULL,
  `updated_by` VARCHAR(100) NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_master_service_categories_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `therapists` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code` VARCHAR(100) NOT NULL,
  `profile` TEXT NULL,
  `name` VARCHAR(255) NOT NULL,
  `is_verified` TINYINT NOT NULL DEFAULT 0,
  `city_id` BIGINT UNSIGNED NULL DEFAULT NULL,
  `experience_year` INT NOT NULL DEFAULT 0,
  `rating` DECIMAL(5,2) NOT NULL DEFAULT 0.00,
  `price` INT NOT NULL DEFAULT 0,
  `auth_id` BIGINT UNSIGNED NOT NULL,
  `therapy_id` BIGINT UNSIGNED NOT NULL,
  `created_at` DATETIME(3) NULL DEFAULT NULL,
  `updated_at` DATETIME(3) NULL DEFAULT NULL,
  `created_by` VARCHAR(100) NULL DEFAULT NULL,
  `updated_by` VARCHAR(100) NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_therapists_code` (`code`),
  KEY `idx_therapists_city_id` (`city_id`),
  KEY `idx_therapists_auth_id` (`auth_id`),
  KEY `idx_therapists_therapy_id` (`therapy_id`),
  KEY `idx_therapists_is_verified` (`is_verified`),
  CONSTRAINT `fk_therapists_city_id` FOREIGN KEY (`city_id`) REFERENCES `master_cities` (`id`),
  CONSTRAINT `fk_therapists_auth_id` FOREIGN KEY (`auth_id`) REFERENCES `user_admins` (`id`),
  CONSTRAINT `fk_therapists_therapy_id` FOREIGN KEY (`therapy_id`) REFERENCES `master_therapy_types` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `bookings` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code` VARCHAR(100) NULL DEFAULT NULL,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `therapist_id` BIGINT UNSIGNED NOT NULL,
  `city_id` BIGINT UNSIGNED NOT NULL,
  `status_id` BIGINT UNSIGNED NOT NULL,
  `date_time` DATETIME(3) NOT NULL,
  `status` VARCHAR(50) NOT NULL,
  `ref_number` VARCHAR(100) NULL DEFAULT NULL,
  `created_at` DATETIME(3) NULL DEFAULT NULL,
  `updated_at` DATETIME(3) NULL DEFAULT NULL,
  `created_by` VARCHAR(100) NULL DEFAULT NULL,
  `updated_by` VARCHAR(100) NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_bookings_code` (`code`),
  KEY `idx_bookings_ref_number` (`ref_number`),
  KEY `idx_bookings_user_id` (`user_id`),
  KEY `idx_bookings_therapist_id` (`therapist_id`),
  KEY `idx_bookings_city_id` (`city_id`),
  KEY `idx_bookings_status_id` (`status_id`),
  KEY `idx_bookings_status` (`status`),
  KEY `idx_bookings_date_time` (`date_time`),
  CONSTRAINT `fk_bookings_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_bookings_therapist_id` FOREIGN KEY (`therapist_id`) REFERENCES `therapists` (`id`),
  CONSTRAINT `fk_bookings_city_id` FOREIGN KEY (`city_id`) REFERENCES `master_cities` (`id`),
  CONSTRAINT `fk_bookings_status_id` FOREIGN KEY (`status_id`) REFERENCES `master_booking_statuses` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `booking_status_histories` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `booking_id` BIGINT UNSIGNED NOT NULL,
  `status_id` BIGINT UNSIGNED NOT NULL,
  `notes` TEXT NULL,
  `created_at` DATETIME(3) NULL DEFAULT NULL,
  `updated_at` DATETIME(3) NULL DEFAULT NULL,
  `created_by` VARCHAR(100) NULL DEFAULT NULL,
  `updated_by` VARCHAR(100) NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_booking_status_histories_booking_id` (`booking_id`),
  KEY `idx_booking_status_histories_status_id` (`status_id`),
  CONSTRAINT `fk_booking_status_histories_booking_id` FOREIGN KEY (`booking_id`) REFERENCES `bookings` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_booking_status_histories_status_id` FOREIGN KEY (`status_id`) REFERENCES `master_booking_statuses` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `payments` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `total` DECIMAL(15,2) NOT NULL DEFAULT 0.00,
  `reference_number` VARCHAR(100) NOT NULL,
  `booking_id` BIGINT UNSIGNED NOT NULL,
  `method` VARCHAR(100) NOT NULL,
  `status` VARCHAR(50) NOT NULL,
  `third_party_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `created_at` DATETIME(3) NULL DEFAULT NULL,
  `updated_at` DATETIME(3) NULL DEFAULT NULL,
  `created_by` VARCHAR(100) NULL DEFAULT NULL,
  `updated_by` VARCHAR(100) NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_payments_reference_number` (`reference_number`),
  KEY `idx_payments_booking_id` (`booking_id`),
  KEY `idx_payments_status` (`status`),
  CONSTRAINT `fk_payments_booking_id` FOREIGN KEY (`booking_id`) REFERENCES `bookings` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `payment_details` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `amount` DECIMAL(15,2) NOT NULL DEFAULT 0.00,
  `reference_number` VARCHAR(100) NOT NULL,
  `name` VARCHAR(100) NOT NULL,
  `parent_payment_id` BIGINT UNSIGNED NOT NULL,
  `created_at` DATETIME(3) NULL DEFAULT NULL,
  `updated_at` DATETIME(3) NULL DEFAULT NULL,
  `created_by` VARCHAR(100) NULL DEFAULT NULL,
  `updated_by` VARCHAR(100) NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_payment_details_reference_number` (`reference_number`),
  KEY `idx_payment_details_parent_payment_id` (`parent_payment_id`),
  CONSTRAINT `fk_payment_details_parent_payment_id` FOREIGN KEY (`parent_payment_id`) REFERENCES `payments` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `services` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(255) NOT NULL,
  `category_id` BIGINT UNSIGNED NOT NULL,
  `description` TEXT NULL,
  `duration` INT NOT NULL DEFAULT 0,
  `price` DECIMAL(15,2) NOT NULL DEFAULT 0.00,
  `commission` DECIMAL(5,2) NOT NULL DEFAULT 0.00,
  `created_at` DATETIME(3) NULL DEFAULT NULL,
  `updated_at` DATETIME(3) NULL DEFAULT NULL,
  `created_by` VARCHAR(100) NULL DEFAULT NULL,
  `updated_by` VARCHAR(100) NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_services_category_id` (`category_id`),
  KEY `idx_services_name` (`name`),
  CONSTRAINT `fk_services_category_id` FOREIGN KEY (`category_id`) REFERENCES `master_service_categories` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `service_areas` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `service_id` BIGINT UNSIGNED NOT NULL,
  `city_id` BIGINT UNSIGNED NOT NULL,
  `created_at` DATETIME(3) NULL DEFAULT NULL,
  `updated_at` DATETIME(3) NULL DEFAULT NULL,
  `created_by` VARCHAR(100) NULL DEFAULT NULL,
  `updated_by` VARCHAR(100) NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_service_areas_service_city` (`service_id`, `city_id`),
  KEY `idx_service_areas_city_id` (`city_id`),
  CONSTRAINT `fk_service_areas_service_id` FOREIGN KEY (`service_id`) REFERENCES `services` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_service_areas_city_id` FOREIGN KEY (`city_id`) REFERENCES `master_cities` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `service_included_items` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `service_id` BIGINT UNSIGNED NOT NULL,
  `name` VARCHAR(255) NOT NULL,
  `created_at` DATETIME(3) NULL DEFAULT NULL,
  `updated_at` DATETIME(3) NULL DEFAULT NULL,
  `created_by` VARCHAR(100) NULL DEFAULT NULL,
  `updated_by` VARCHAR(100) NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_service_included_items_service_id` (`service_id`),
  CONSTRAINT `fk_service_included_items_service_id` FOREIGN KEY (`service_id`) REFERENCES `services` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `settings` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `application_fee` DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  `tax_ppn` DECIMAL(5,2) NOT NULL DEFAULT 0.00,
  `service_fee` DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  `status` TINYINT NOT NULL DEFAULT 1,
  `created_at` DATETIME(3) NULL DEFAULT NULL,
  `updated_at` DATETIME(3) NULL DEFAULT NULL,
  `created_by` VARCHAR(100) NULL DEFAULT NULL,
  `updated_by` VARCHAR(100) NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_settings_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `summary_homepage` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `terapis` INT NOT NULL DEFAULT 0,
  `kota` INT NOT NULL DEFAULT 0,
  `rating` INT NOT NULL DEFAULT 0,
  `created_at` DATETIME(3) NULL DEFAULT NULL,
  `updated_at` DATETIME(3) NULL DEFAULT NULL,
  `created_by` VARCHAR(100) NULL DEFAULT NULL,
  `updated_by` VARCHAR(100) NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

SET FOREIGN_KEY_CHECKS = 1;

INSERT INTO `master_roles` (`id`, `name`, `label`, `created_at`, `updated_at`, `created_by`, `updated_by`) VALUES
  (1, 'SUPERADMIN', 'Super Admin', CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3), 'system', 'system'),
  (2, 'ADMIN', 'Admin', CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3), 'system', 'system'),
  (3, 'TERAPIS', 'Terapis', CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3), 'system', 'system'),
  (4, 'USER', 'User', CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3), 'system', 'system')
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `label` = VALUES(`label`),
  `updated_at` = CURRENT_TIMESTAMP(3),
  `updated_by` = 'system';

INSERT INTO `master_booking_statuses` (`name`, `code`, `created_at`, `updated_at`, `created_by`, `updated_by`) VALUES
  ('pending', 'PENDING', CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3), 'system', 'system'),
  ('scheduled', 'SCHEDULED', CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3), 'system', 'system'),
  ('completed', 'COMPLETED', CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3), 'system', 'system'),
  ('canceled', 'CANCELED', CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3), 'system', 'system')
ON DUPLICATE KEY UPDATE
  `code` = VALUES(`code`),
  `updated_at` = CURRENT_TIMESTAMP(3),
  `updated_by` = 'system';

INSERT INTO `master_cities` (`id`, `name`, `created_at`, `updated_at`, `created_by`, `updated_by`) VALUES
  (1, 'Jakarta', CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3), 'system', 'system'),
  (2, 'Bandung', CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3), 'system', 'system'),
  (3, 'Surabaya', CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3), 'system', 'system'),
  (4, 'Yogyakarta', CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3), 'system', 'system'),
  (5, 'Denpasar', CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3), 'system', 'system')
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `updated_at` = CURRENT_TIMESTAMP(3),
  `updated_by` = 'system';

INSERT INTO `master_therapy_types` (`id`, `name`, `created_at`, `updated_at`, `created_by`, `updated_by`) VALUES
  (1, 'Physiotherapy', CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3), 'system', 'system')
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `updated_at` = CURRENT_TIMESTAMP(3),
  `updated_by` = 'system';

INSERT INTO `user_admins` (`id`, `code`, `full_name`, `email`, `phone`, `role_id`, `password`, `auth_code`, `is_active`, `chat_ref`, `created_at`, `updated_at`, `created_by`, `updated_by`) VALUES
  (1001, 'TRP-DR-001', 'Dr. Andi Pratama', 'therapist.doctor1@phisiobook.local', '081200000001', 3, '', NULL, 1, 'THERAPIST-DR-001', CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3), 'system', 'system'),
  (1002, 'TRP-DR-002', 'Dr. Budi Santoso', 'therapist.doctor2@phisiobook.local', '081200000002', 3, '', NULL, 1, 'THERAPIST-DR-002', CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3), 'system', 'system'),
  (1003, 'TRP-DR-003', 'Dr. Citra Lestari', 'therapist.doctor3@phisiobook.local', '081200000003', 3, '', NULL, 1, 'THERAPIST-DR-003', CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3), 'system', 'system'),
  (1004, 'TRP-DR-004', 'Dr. Dimas Wicaksono', 'therapist.doctor4@phisiobook.local', '081200000004', 3, '', NULL, 1, 'THERAPIST-DR-004', CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3), 'system', 'system'),
  (1005, 'TRP-DR-005', 'Dr. Eka Maharani', 'therapist.doctor5@phisiobook.local', '081200000005', 3, '', NULL, 1, 'THERAPIST-DR-005', CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3), 'system', 'system')
ON DUPLICATE KEY UPDATE
  `code` = VALUES(`code`),
  `full_name` = VALUES(`full_name`),
  `phone` = VALUES(`phone`),
  `role_id` = VALUES(`role_id`),
  `is_active` = VALUES(`is_active`),
  `chat_ref` = VALUES(`chat_ref`),
  `updated_at` = CURRENT_TIMESTAMP(3),
  `updated_by` = 'system';

INSERT INTO `therapists` (`id`, `code`, `profile`, `name`, `is_verified`, `city_id`, `experience_year`, `rating`, `price`, `auth_id`, `therapy_id`, `created_at`, `updated_at`, `created_by`, `updated_by`) VALUES
  (1001, 'THR-001', NULL, 'Dr. Andi Pratama', 1, 1, 8, 4.80, 250000, 1001, 1, CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3), 'system', 'system'),
  (1002, 'THR-002', NULL, 'Dr. Budi Santoso', 1, 2, 7, 4.70, 240000, 1002, 1, CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3), 'system', 'system'),
  (1003, 'THR-003', NULL, 'Dr. Citra Lestari', 1, 3, 9, 4.90, 275000, 1003, 1, CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3), 'system', 'system'),
  (1004, 'THR-004', NULL, 'Dr. Dimas Wicaksono', 1, 4, 6, 4.60, 225000, 1004, 1, CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3), 'system', 'system'),
  (1005, 'THR-005', NULL, 'Dr. Eka Maharani', 1, 5, 10, 4.95, 300000, 1005, 1, CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3), 'system', 'system')
ON DUPLICATE KEY UPDATE
  `profile` = VALUES(`profile`),
  `name` = VALUES(`name`),
  `is_verified` = VALUES(`is_verified`),
  `city_id` = VALUES(`city_id`),
  `experience_year` = VALUES(`experience_year`),
  `rating` = VALUES(`rating`),
  `price` = VALUES(`price`),
  `auth_id` = VALUES(`auth_id`),
  `therapy_id` = VALUES(`therapy_id`),
  `updated_at` = CURRENT_TIMESTAMP(3),
  `updated_by` = 'system';

INSERT INTO `settings` (`application_fee`, `tax_ppn`, `service_fee`, `status`, `created_at`, `updated_at`, `created_by`, `updated_by`)
SELECT 0.00, 0.00, 0.00, 1, CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3), 'system', 'system'
WHERE NOT EXISTS (
  SELECT 1 FROM `settings` WHERE `status` = 1
);

INSERT INTO `summary_homepage` (`id`, `terapis`, `kota`, `rating`, `created_at`, `updated_at`, `created_by`, `updated_by`) VALUES
  (1, 5, 5, 5, CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3), 'system', 'system')
ON DUPLICATE KEY UPDATE
  `terapis` = VALUES(`terapis`),
  `kota` = VALUES(`kota`),
  `rating` = VALUES(`rating`),
  `updated_at` = CURRENT_TIMESTAMP(3),
  `updated_by` = 'system';
