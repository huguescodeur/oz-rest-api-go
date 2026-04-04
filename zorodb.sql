-- phpMyAdmin SQL Dump
-- version 5.2.3
-- https://www.phpmyadmin.net/
--
-- Hôte : localhost:8889
-- Généré le : sam. 04 avr. 2026 à 07:34
-- Version du serveur : 8.0.44
-- Version de PHP : 8.3.30

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Base de données : `zorodb`
--

-- --------------------------------------------------------

--
-- Structure de la table `categories`
--

CREATE TABLE `categories` (
  `category_id` int NOT NULL,
  `uuid` char(36) NOT NULL,
  `category_name` varchar(100) NOT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  `user_id` int NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Déchargement des données de la table `categories`
--

INSERT INTO `categories` (`category_id`, `uuid`, `category_name`, `deleted_at`, `user_id`) VALUES
(1, '263a5cd0-2d2f-11f1-a267-4fa3467576ef', 'Informatique', NULL, 2),
(2, '263a5e74-2d2f-11f1-a267-4fa3467576ef', 'Bureautique', NULL, 1),
(3, '263a5ec4-2d2f-11f1-a267-4fa3467576ef', 'Accessoires', NULL, 2);

-- --------------------------------------------------------

--
-- Structure de la table `orders`
--

CREATE TABLE `orders` (
  `order_id` int NOT NULL,
  `uuid` char(36) NOT NULL,
  `shop_id` int DEFAULT NULL,
  `user_id` int NOT NULL,
  `total_amount` int DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Structure de la table `order_items`
--

CREATE TABLE `order_items` (
  `item_id` int NOT NULL,
  `order_id` int DEFAULT NULL,
  `product_id` int DEFAULT NULL,
  `quantity` int DEFAULT NULL,
  `unit_price` int DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Structure de la table `products`
--

CREATE TABLE `products` (
  `product_id` int NOT NULL,
  `uuid` char(36) NOT NULL,
  `product_name` varchar(255) NOT NULL,
  `unit_price` int NOT NULL,
  `category_id` int NOT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` timestamp NULL DEFAULT NULL,
  `user_id` int NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Déchargement des données de la table `products`
--

INSERT INTO `products` (`product_id`, `uuid`, `product_name`, `unit_price`, `category_id`, `created_at`, `updated_at`, `deleted_at`, `user_id`) VALUES
(1, 'c0d25081-45b1-42e2-a95b-46c82cfe8112', 'Clavier Mécanique RGB', 5700, 1, '2026-03-29 06:37:13', '2026-03-31 19:10:26', '2026-03-31 19:10:26', 2),
(2, '263908da-2d2f-11f1-a267-4fa3467576ef', 'Souris Sans Fil', 4500, 3, '2026-03-29 06:37:13', '2026-03-31 19:31:41', NULL, 2),
(3, '2639093e-2d2f-11f1-a267-4fa3467576ef', 'Bureau Assis-Debout', 31000, 2, '2026-03-29 06:37:13', '2026-03-31 18:26:35', NULL, 1),
(4, '2639097a-2d2f-11f1-a267-4fa3467576ef', 'Écran 27 pouces 4K', 29900, 1, '2026-03-29 06:37:13', '2026-03-31 18:26:35', NULL, 2),
(5, '26390a06-2d2f-11f1-a267-4fa3467576ef', 'Tapis de souris XL', 1500, 3, '2026-03-29 06:37:13', '2026-03-31 18:26:35', NULL, 2),
(9, '26390a42-2d2f-11f1-a267-4fa3467576ef', 'Clavier Lumineux', 1800, 1, '2026-03-29 17:52:30', '2026-03-31 18:26:35', NULL, 2),
(11, '26390a74-2d2f-11f1-a267-4fa3467576ef', 'Clavier Lumineux', 2200, 3, '2026-03-30 12:35:31', '2026-03-31 18:26:35', NULL, 2),
(12, '26390a9c-2d2f-11f1-a267-4fa3467576ef', 'Clavier Lumineux', 2800, 3, '2026-03-30 12:41:38', '2026-03-31 18:26:35', NULL, 2),
(13, '26390ace-2d2f-11f1-a267-4fa3467576ef', 'Chargeur PC HP', 3200, 2, '2026-03-31 01:53:03', '2026-03-31 18:26:35', NULL, 1),
(14, '26390af6-2d2f-11f1-a267-4fa3467576ef', 'Chargeur PC DELL', 15500, 2, '2026-03-31 02:03:59', '2026-03-31 18:26:35', NULL, 1),
(15, '26390b1e-2d2f-11f1-a267-4fa3467576ef', 'Chargeur PC LENOVO', 15500, 2, '2026-03-31 02:08:49', '2026-03-31 18:26:35', NULL, 1),
(16, '26390b46-2d2f-11f1-a267-4fa3467576ef', 'Casque Gamer R', 35500, 2, '2026-03-31 02:12:33', '2026-03-31 18:26:35', NULL, 1),
(17, '26390b6e-2d2f-11f1-a267-4fa3467576ef', 'Casque Gamer R', 37500, 2, '2026-03-31 02:17:07', '2026-03-31 18:26:35', NULL, 1),
(18, '26390b96-2d2f-11f1-a267-4fa3467576ef', 'Casque Gamer X', 38600, 2, '2026-03-31 02:19:54', '2026-03-31 18:26:35', NULL, 1),
(19, '26390bc8-2d2f-11f1-a267-4fa3467576ef', 'Casque Gamer X', 38600, 3, '2026-03-31 02:49:05', '2026-03-31 18:26:35', NULL, 2),
(20, '26390be6-2d2f-11f1-a267-4fa3467576ef', 'Casque Gamer LT', 40600, 2, '2026-03-31 02:49:24', '2026-03-31 18:26:35', NULL, 1),
(21, '26390c0e-2d2f-11f1-a267-4fa3467576ef', 'Casque Gamer LT', 40600, 2, '2026-03-31 14:45:41', '2026-03-31 18:26:35', NULL, 0),
(22, '26390c36-2d2f-11f1-a267-4fa3467576ef', 'Casque Gamer LT', 0, 2, '2026-03-31 14:58:26', '2026-03-31 18:26:35', NULL, 0),
(23, '26390c5e-2d2f-11f1-a267-4fa3467576ef', 'Casque Gamer LT', 234, 2, '2026-03-31 15:10:10', '2026-03-31 18:26:35', NULL, 1),
(24, '26390c86-2d2f-11f1-a267-4fa3467576ef', 'Casque Gamer LT', 2700, 2, '2026-03-31 15:12:02', '2026-04-01 07:17:27', NULL, 1),
(25, '1a3a93b1-11f6-41ac-9f8e-bd890e876ab2', 'Souris avec fil', 1200, 3, '2026-03-31 20:18:04', '2026-03-31 20:29:04', NULL, 2),
(26, 'cd165466-64e4-4e1d-a4e8-84519a623eaf', 'Maths', 1000, 1, '2026-04-02 14:42:04', '2026-04-02 14:42:04', NULL, 20);

-- --------------------------------------------------------

--
-- Structure de la table `shops`
--

CREATE TABLE `shops` (
  `shop_id` int NOT NULL,
  `uuid` char(36) NOT NULL,
  `shop_name` varchar(100) NOT NULL,
  `shop_address` varchar(255) NOT NULL,
  `shop_phone` varchar(20) DEFAULT NULL,
  `shop_mail` varchar(150) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` timestamp NULL DEFAULT NULL,
  `user_id` int NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Déchargement des données de la table `shops`
--

INSERT INTO `shops` (`shop_id`, `uuid`, `shop_name`, `shop_address`, `shop_phone`, `shop_mail`, `created_at`, `updated_at`, `deleted_at`, `user_id`) VALUES
(1, '895fe6aa-2d21-11f1-a267-4fa3467576ef', 'Chic Shop Yop', 'Yopougon Rond Point Gandhi', '0598675460', 'chicshop@gmail.com', '2026-03-30 05:45:09', '2026-03-31 23:56:46', NULL, 2),
(3, '217fe77c-2d21-11f1-a267-4fa3467576ef', 'Chic Shop Yop Terminus 40', 'Yopougon Terminus 40', '0510675460', 'chicshop@gmail.com', '2026-03-30 12:11:21', '2026-03-31 16:54:42', NULL, 1),
(4, '3c5c67df-a9a3-4832-bac9-d6b4accfd812', 'NEXIUM.AI Store', 'Yopougon Camp-Militaire', '6028675460', 'ai.nexium@gmail.com', '2026-03-31 23:59:26', '2026-03-31 23:59:26', NULL, 2);

-- --------------------------------------------------------

--
-- Structure de la table `stocks`
--

CREATE TABLE `stocks` (
  `product_id` int NOT NULL,
  `shop_id` int NOT NULL,
  `quantity` int DEFAULT '0',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Structure de la table `stock_movements`
--

CREATE TABLE `stock_movements` (
  `movement_id` int NOT NULL,
  `product_id` int NOT NULL,
  `shop_id` int NOT NULL,
  `user_id` int NOT NULL,
  `quantity_change` int NOT NULL,
  `movement_type` enum('SALE','STOCK_IN','GIFT','LOSS','ADJUSTMENT') NOT NULL,
  `comment` varchar(255) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Structure de la table `users`
--

CREATE TABLE `users` (
  `id` int NOT NULL,
  `uuid` char(36) NOT NULL,
  `username` varchar(50) NOT NULL,
  `firstname` varchar(100) NOT NULL,
  `lastname` varchar(100) NOT NULL,
  `email` varchar(150) NOT NULL,
  `phone` varchar(20) DEFAULT NULL,
  `password_hash` varchar(255) NOT NULL,
  `role` enum('super','admin','vendeur') CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'vendeur',
  `profile_picture` varchar(255) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` timestamp NULL DEFAULT NULL,
  `parent_id` int DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Déchargement des données de la table `users`
--

INSERT INTO `users` (`id`, `uuid`, `username`, `firstname`, `lastname`, `email`, `phone`, `password_hash`, `role`, `profile_picture`, `created_at`, `updated_at`, `deleted_at`, `parent_id`) VALUES
(1, '29e2afe5-50b9-46c0-a6ec-b30b6a36b5cc', 'hgsdev', 'Hgs', 'Dev', 'hgsdev@example.com', '0503020385', '$2a$10$Q0qY2f7Axq0mX0nMPETurOV176GySWP5NYyMPn50JRyVU/FJjOl3K', 'admin', NULL, '2026-04-01 04:35:51', '2026-04-01 05:57:35', NULL, NULL),
(2, '4fce273d-b5c0-40cf-ad32-d38ffe5bdaf5', 'zoro_dev', 'Roronoa', 'Zoro', 'zoro@example.com', '0123456789', '$2a$10$sYAUcLuqOv2UxeAdn5KFTemOe7lKJz4r1ESlj8dLtaYxzVB6jQt2G', 'vendeur', NULL, '2026-04-01 06:01:24', '2026-04-01 06:01:24', NULL, 1),
(4, 'e072fe98-824b-49cf-be80-1c0b32c0d9e1', 'gnabacharles', 'Charles', 'Gnaba', 'gc@example.com', '0154632454', '$2a$10$8vVeyo0fZ5eFAbr8cwUaBOUI57CmNf7pk8q..1BQi.3LJ6xlw8/Pm', 'vendeur', NULL, '2026-04-01 07:06:17', '2026-04-03 13:23:54', NULL, 1),
(5, '9028bcae-2ce0-4289-a095-686ab41c1b9f', 'kpriscille', 'Priscille', 'Kouamé', 'pk@example.com', '0700632454', '$2a$10$sM9Hou9KiqNMKwYKjULGDOvLBA5R/VgjHEVK7Qqb6a9SxhPKAEc4q', 'vendeur', NULL, '2026-04-01 07:14:31', '2026-04-01 07:14:31', NULL, 1),
(6, 'a9780f12-a77b-4715-8ea1-49e81d57e82f', 'yannsela', 'Yann', 'Sela', 'ys@example.com', '0554632400', '$2a$10$SIDRZXNbfdezzuv.28BMketD4HCxRxCD6vP34zzqe4/PGWl0BWkhS', 'vendeur', NULL, '2026-04-01 07:30:10', '2026-04-01 07:30:10', NULL, 1),
(7, '317e8092-e67f-4e69-bb2c-fa8bca3e6105', 'admin2', 'Two', 'Admin', 'at@example.com', '0554632454', '$2a$10$rOewy/APKX16PMbon0IupuH09EcYt.tUS2h3NhLN8P8aciWzGjUYq', 'admin', NULL, '2026-04-01 21:27:52', '2026-04-03 13:32:41', NULL, NULL),
(8, '8f1053fd-0445-46d5-9385-7e3446ccd080', 'kenpat', 'Pat', 'Ken', 'kp@example.com', '0554630051', '$2a$10$nzXRRoYFMdn1xg5RMVhdpOp7Slhh/WWuHkGzvJLuXpKjts0KP9o0y', 'vendeur', NULL, '2026-04-01 21:30:11', '2026-04-01 21:30:11', NULL, 7),
(9, '73467ba3-e6b0-4bc3-803d-f1fc43c9433d', 'codeclaude', 'Code', 'Claude', 'c@example.com', '0554630051', '$2a$10$Qq1lRjqKnyRLmXyNKKhBD.YYiHCWc9ots/rWJPYHEHa5FWi1iMV3a', 'admin', NULL, '2026-04-01 22:45:18', '2026-04-03 12:07:12', NULL, NULL),
(10, '1af6a38b-d635-46ec-8a86-bda5ae51727a', 'angekonan', 'Ange', 'Konan', 'kn@example.com', '0554630051', '$2a$10$xfMshxj3tkX4TH4yBRlWe.E8T4aA6mOGPSiw5A5Tc4G42NIzgtR6q', 'vendeur', NULL, '2026-04-01 22:46:21', '2026-04-01 22:46:21', NULL, 9),
(13, '12eaccb1-5f09-41d8-b862-fed6e83272ff', 'kaderkonan', 'Kader', 'Konan', 'kk@example.com', '0554630051', '$2a$10$Cc5fvlBNirvbFv6Z3EOsPOBB7vHIK69g356HwZUZnXu/gbvfRiB2y', 'admin', NULL, '2026-04-01 22:51:55', '2026-04-01 22:51:55', NULL, NULL),
(15, 'c632d78d-eff6-4738-a816-fee894ffa07b', 'rpkonan', 'Prince', 'Konan', 'prk@example.com', '0554630051', '$2a$10$DsW3./XQVlz7Ijf0YMM6OuMHEkrFpF2wCpqRrhFcxGr9n.0f77fni', 'vendeur', NULL, '2026-04-01 22:53:01', '2026-04-02 03:23:19', NULL, 13),
(16, 'eb3c9af3-f4ef-43f5-ac97-dfae690c9e19', 'ismo', 'Ismo', 'Kader', 'is@example.com', '0554630051', '$2a$10$AfPrpixCKJJsLgxBGSoLN.QG7ZBcx2ij9eXVqiwLdM9wq8uu17joq', 'admin', NULL, '2026-04-02 04:52:33', '2026-04-02 04:58:04', NULL, NULL),
(17, '71fcca14-a1bc-4b6b-8e9c-7a6769279498', 'fabriceirie', 'Fabrice', 'Irie', 'if@example.com', '0554630051', '$2a$10$FdENx5HAKXtoZ8O5whGQuO4rdN4FDTdeqAvcBHtVHgISrWOWthURe', 'admin', NULL, '2026-04-02 05:06:15', '2026-04-02 05:06:15', NULL, NULL),
(18, '98336598-4f3e-46c7-a440-2b7e137fafc2', 'jeand7', 'Jean', 'Satchi', 'sj@example.com', '0554630051', '$2a$10$wzBhVHPx7bZKiyzgJO2PTu1Ijiz9OArdc8zlXsB0/V.m50InBB7ei', 'admin', NULL, '2026-04-02 05:20:14', '2026-04-02 05:54:37', NULL, NULL),
(19, '0b29d03a-71b5-4a06-a1f0-cac408a85a4f', 'huguescodeur', 'Hugues', 'Codeur', 'hg@example.com', '0503020385', '$2a$10$ZXBFE.kU3wsdYymnfksJaeUo/rLffd8tx3aoMpnMb0k.1Wvv9eWea', 'super', NULL, '2026-04-02 05:57:18', '2026-04-02 05:57:18', NULL, NULL),
(20, 'f7ae417f-588c-4a7a-b78a-5b3b5870038e', 'dar_ryl', 'Daryl', 'DEGAN', 'degandaryl@gmail.com', '0169349151', '$2a$10$nBE4zM/MK6ph5e/TzsaPAu83NGEFR2Dcw4evCaM.rRjZGzRhvEeS.', 'admin', NULL, '2026-04-02 13:04:40', '2026-04-02 13:04:40', NULL, NULL),
(21, '08c3debd-5e1d-444d-89a2-8f4c8bd257fc', 'kaderkouadio', 'Kader', 'Kouadio', 'kk@gmail.com', '0503020385', '$2a$10$YFClCztdX1mAJbm5uviQGuGdn4jSvDbf/9dtf/FXYMo49gfdmCOyW', 'vendeur', NULL, '2026-04-03 08:58:15', '2026-04-03 08:58:15', NULL, NULL),
(22, '5f1a7087-8854-401d-82ae-84c0a4698560', 'gnabacharles09', 'c&', 'g&', 'gc&@example.com', '0154632454', '$2a$10$QLtEIw1w6Kfee9Fmt4BaQOviyuOJExCJmQUCnooRJiGvLV1mPOfFm', 'vendeur', NULL, '2026-04-03 23:22:22', '2026-04-03 23:22:22', NULL, NULL),
(23, '177b9e47-df9e-4df3-990e-ae7b53346ce6', 'gc12', 'c&', 'g&', 'gc12@gmail.com', '0754632454', '$2a$10$6anythJu9FoV2BLIUuk/lOoS4OXe4fcAeiaoCcWcPCBIUha1tFAFi', 'vendeur', NULL, '2026-04-03 23:31:49', '2026-04-03 23:31:49', NULL, NULL),
(24, 'be65684a-c44d-44d9-8168-42604b55324a', 'baz', 'b', 'a', 'baz@gmail.com', '0554630051', '$2a$10$1lqF.DN83Ct/8Bnio55Rbene2h3/WVCcRHwhyft5KCw2VE4uslTcO', 'vendeur', NULL, '2026-04-03 23:39:30', '2026-04-03 23:39:30', NULL, NULL),
(26, '41eb5d51-4332-4826-a54f-30adea1cb233', 'adamo2', 'ad', 'mo', 'adm@gmail.com', '0154632454', '$2a$10$Y9OgSHTqP7727TPxC8XAYuO6FcHjvCVXv6rAhJxdrc0PQNQl/kufi', 'admin', NULL, '2026-04-04 00:59:46', '2026-04-04 00:59:46', NULL, NULL);

-- --------------------------------------------------------

--
-- Structure de la table `user_shops`
--

CREATE TABLE `user_shops` (
  `user_id` int NOT NULL,
  `shop_id` int NOT NULL,
  `assigned_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Déchargement des données de la table `user_shops`
--

INSERT INTO `user_shops` (`user_id`, `shop_id`, `assigned_at`) VALUES
(2, 3, '2026-04-03 02:46:38');

--
-- Index pour les tables déchargées
--

--
-- Index pour la table `categories`
--
ALTER TABLE `categories`
  ADD PRIMARY KEY (`category_id`),
  ADD UNIQUE KEY `category_name` (`category_name`),
  ADD UNIQUE KEY `idx_categories_uuid` (`uuid`);

--
-- Index pour la table `orders`
--
ALTER TABLE `orders`
  ADD PRIMARY KEY (`order_id`),
  ADD UNIQUE KEY `idx_orders_uuid` (`uuid`),
  ADD KEY `shop_id` (`shop_id`),
  ADD KEY `user_id` (`user_id`);

--
-- Index pour la table `order_items`
--
ALTER TABLE `order_items`
  ADD PRIMARY KEY (`item_id`),
  ADD KEY `order_id` (`order_id`),
  ADD KEY `product_id` (`product_id`);

--
-- Index pour la table `products`
--
ALTER TABLE `products`
  ADD PRIMARY KEY (`product_id`),
  ADD UNIQUE KEY `idx_products_uuid` (`uuid`),
  ADD KEY `fk_category` (`category_id`),
  ADD KEY `idx_products_user` (`user_id`),
  ADD KEY `idx_categories_user` (`user_id`),
  ADD KEY `idx_orders_user` (`user_id`);

--
-- Index pour la table `shops`
--
ALTER TABLE `shops`
  ADD PRIMARY KEY (`shop_id`),
  ADD UNIQUE KEY `idx_shops_uuid` (`uuid`),
  ADD KEY `idx_shops_user` (`user_id`);

--
-- Index pour la table `stocks`
--
ALTER TABLE `stocks`
  ADD PRIMARY KEY (`product_id`,`shop_id`),
  ADD UNIQUE KEY `unique_stock` (`product_id`,`shop_id`),
  ADD KEY `shop_id` (`shop_id`);

--
-- Index pour la table `stock_movements`
--
ALTER TABLE `stock_movements`
  ADD PRIMARY KEY (`movement_id`),
  ADD KEY `product_id` (`product_id`),
  ADD KEY `shop_id` (`shop_id`),
  ADD KEY `idx_movements_user` (`user_id`);

--
-- Index pour la table `users`
--
ALTER TABLE `users`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `username` (`username`),
  ADD UNIQUE KEY `email` (`email`),
  ADD UNIQUE KEY `idx_users_uuid` (`uuid`),
  ADD KEY `idx_users_parent` (`parent_id`);

--
-- Index pour la table `user_shops`
--
ALTER TABLE `user_shops`
  ADD PRIMARY KEY (`user_id`,`shop_id`),
  ADD KEY `fk_shop_id` (`shop_id`);

--
-- AUTO_INCREMENT pour les tables déchargées
--

--
-- AUTO_INCREMENT pour la table `categories`
--
ALTER TABLE `categories`
  MODIFY `category_id` int NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=4;

--
-- AUTO_INCREMENT pour la table `orders`
--
ALTER TABLE `orders`
  MODIFY `order_id` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `order_items`
--
ALTER TABLE `order_items`
  MODIFY `item_id` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `products`
--
ALTER TABLE `products`
  MODIFY `product_id` int NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=27;

--
-- AUTO_INCREMENT pour la table `shops`
--
ALTER TABLE `shops`
  MODIFY `shop_id` int NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=5;

--
-- AUTO_INCREMENT pour la table `stock_movements`
--
ALTER TABLE `stock_movements`
  MODIFY `movement_id` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `users`
--
ALTER TABLE `users`
  MODIFY `id` int NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=27;

--
-- Contraintes pour les tables déchargées
--

--
-- Contraintes pour la table `orders`
--
ALTER TABLE `orders`
  ADD CONSTRAINT `orders_ibfk_1` FOREIGN KEY (`shop_id`) REFERENCES `shops` (`shop_id`),
  ADD CONSTRAINT `orders_ibfk_2` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`);

--
-- Contraintes pour la table `order_items`
--
ALTER TABLE `order_items`
  ADD CONSTRAINT `order_items_ibfk_1` FOREIGN KEY (`order_id`) REFERENCES `orders` (`order_id`),
  ADD CONSTRAINT `order_items_ibfk_2` FOREIGN KEY (`product_id`) REFERENCES `products` (`product_id`);

--
-- Contraintes pour la table `products`
--
ALTER TABLE `products`
  ADD CONSTRAINT `fk_category` FOREIGN KEY (`category_id`) REFERENCES `categories` (`category_id`) ON DELETE RESTRICT;

--
-- Contraintes pour la table `stocks`
--
ALTER TABLE `stocks`
  ADD CONSTRAINT `stocks_ibfk_1` FOREIGN KEY (`product_id`) REFERENCES `products` (`product_id`) ON DELETE CASCADE,
  ADD CONSTRAINT `stocks_ibfk_2` FOREIGN KEY (`shop_id`) REFERENCES `shops` (`shop_id`) ON DELETE CASCADE;

--
-- Contraintes pour la table `stock_movements`
--
ALTER TABLE `stock_movements`
  ADD CONSTRAINT `stock_movements_ibfk_1` FOREIGN KEY (`product_id`) REFERENCES `products` (`product_id`),
  ADD CONSTRAINT `stock_movements_ibfk_2` FOREIGN KEY (`shop_id`) REFERENCES `shops` (`shop_id`),
  ADD CONSTRAINT `stock_movements_ibfk_3` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`);

--
-- Contraintes pour la table `users`
--
ALTER TABLE `users`
  ADD CONSTRAINT `fk_user_parent` FOREIGN KEY (`parent_id`) REFERENCES `users` (`id`) ON DELETE SET NULL;

--
-- Contraintes pour la table `user_shops`
--
ALTER TABLE `user_shops`
  ADD CONSTRAINT `fk_shop_id` FOREIGN KEY (`shop_id`) REFERENCES `shops` (`shop_id`) ON DELETE CASCADE,
  ADD CONSTRAINT `fk_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `fk_user_vendeur` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
