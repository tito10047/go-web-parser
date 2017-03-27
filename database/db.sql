/*
SQLyog Community
MySQL - 5.7.15-log : Database - stavkova
*********************************************************************
*/

/*!40101 SET NAMES utf8 */;

/*!40101 SET SQL_MODE=''*/;

/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;
/*Table structure for table `bet_company` */

DROP TABLE IF EXISTS `bet_company`;

CREATE TABLE `bet_company` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(200) DEFAULT NULL,
  `url` varchar(200) DEFAULT NULL,
  `exchange` bit(1) DEFAULT NULL,
  `routines_count` int(5) unsigned DEFAULT '100',
  `tasks_per_time` int(5) unsigned DEFAULT '60',
  `wait_sec_per_tasks` int(5) unsigned DEFAULT '1',
  `enabled` tinyint(1) unsigned DEFAULT '0',
  `timezone` int(2) DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8;

/*Table structure for table `bet_mach_selection` */

DROP TABLE IF EXISTS `bet_mach_selection`;

CREATE TABLE `bet_mach_selection` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `id_bet_match` int(10) unsigned NOT NULL,
  `name` varchar(255) NOT NULL,
  `odds` float(10,4) NOT NULL,
  `last_update` datetime DEFAULT NULL,
  `org_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `org_id` (`org_id`),
  KEY `all` (`id_bet_match`,`org_id`),
  KEY `id_entry` (`id_bet_match`),
  CONSTRAINT `bet_mach_selection_ibfk_1` FOREIGN KEY (`id_bet_match`) REFERENCES `bet_match` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8;

/*Table structure for table `bet_match` */

DROP TABLE IF EXISTS `bet_match`;

CREATE TABLE `bet_match` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `id_bet_company` int(10) unsigned NOT NULL,
  `id_bet_sport` int(10) unsigned NOT NULL,
  `id_bet_match_type` int(10) unsigned NOT NULL,
  `name` varchar(255) NOT NULL,
  `team_a` int(10) unsigned DEFAULT NULL,
  `team_b` int(10) unsigned DEFAULT NULL,
  `date` datetime NOT NULL,
  `org_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `id_bet_sport` (`id_bet_sport`),
  KEY `date` (`date`),
  KEY `org_id` (`org_id`),
  KEY `id_bet_match_type` (`id_bet_match_type`),
  KEY `team_a` (`team_a`),
  KEY `team_b` (`team_b`),
  KEY `all` (`id_bet_company`,`id_bet_sport`,`id_bet_match_type`,`org_id`),
  CONSTRAINT `bet_match_ibfk_1` FOREIGN KEY (`id_bet_company`) REFERENCES `bet_company` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `bet_match_ibfk_2` FOREIGN KEY (`id_bet_sport`) REFERENCES `bet_sport` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `bet_match_ibfk_3` FOREIGN KEY (`id_bet_match_type`) REFERENCES `bet_match_type` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `bet_match_ibfk_4` FOREIGN KEY (`team_a`) REFERENCES `bet_team` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `bet_match_ibfk_5` FOREIGN KEY (`team_b`) REFERENCES `bet_team` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;

/*Table structure for table `bet_match_type` */

DROP TABLE IF EXISTS `bet_match_type`;

CREATE TABLE `bet_match_type` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(200) DEFAULT NULL,
  `enabled` tinyint(1) DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8;

/*Table structure for table `bet_match_type_name` */

DROP TABLE IF EXISTS `bet_match_type_name`;

CREATE TABLE `bet_match_type_name` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `id_bet_match_type` int(10) unsigned DEFAULT NULL,
  `name` varchar(100) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `id_bet_match_type` (`id_bet_match_type`),
  CONSTRAINT `bet_match_type_name_ibfk_1` FOREIGN KEY (`id_bet_match_type`) REFERENCES `bet_match_type` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

/*Table structure for table `bet_sport` */

DROP TABLE IF EXISTS `bet_sport`;

CREATE TABLE `bet_sport` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(200) DEFAULT NULL,
  `enabled` tinyint(1) DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8;

/*Table structure for table `bet_sport_name` */

DROP TABLE IF EXISTS `bet_sport_name`;

CREATE TABLE `bet_sport_name` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `id_bet_sport` int(10) unsigned DEFAULT NULL,
  `name` varchar(100) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`),
  KEY `id_bet_sport` (`id_bet_sport`),
  CONSTRAINT `bet_sport_name_ibfk_1` FOREIGN KEY (`id_bet_sport`) REFERENCES `bet_sport` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=78 DEFAULT CHARSET=utf8;

/*Table structure for table `bet_team` */

DROP TABLE IF EXISTS `bet_team`;

CREATE TABLE `bet_team` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `id_bet_sport` int(10) unsigned NOT NULL,
  `name` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `id_bet_sport` (`id_bet_sport`),
  CONSTRAINT `bet_team_ibfk_1` FOREIGN KEY (`id_bet_sport`) REFERENCES `bet_sport` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=1105 DEFAULT CHARSET=utf8;

/*Table structure for table `bet_team_name` */

DROP TABLE IF EXISTS `bet_team_name`;

CREATE TABLE `bet_team_name` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `id_bet_sport` int(10) unsigned DEFAULT NULL,
  `id_bet_team` int(10) unsigned DEFAULT NULL,
  `name` varchar(200) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_bet_sport` (`id_bet_sport`,`name`),
  KEY `id_bet_team` (`id_bet_team`),
  CONSTRAINT `bet_team_name_ibfk_1` FOREIGN KEY (`id_bet_sport`) REFERENCES `bet_sport` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `bet_team_name_ibfk_2` FOREIGN KEY (`id_bet_team`) REFERENCES `bet_team` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=562 DEFAULT CHARSET=utf8;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
