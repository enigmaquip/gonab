/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `category` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) DEFAULT NULL,
  `active` tinyint(1) DEFAULT NULL,
  `description` varchar(1024) DEFAULT NULL,
  `disable_preview` tinyint(1) DEFAULT NULL,
  `min_size` int(11) DEFAULT NULL,
  `parent_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
);
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `category` VALUES (0,'Other',1,NULL,0,0,NULL),(10,'Misc',1,NULL,0,0,0),(20,'Hashed',1,NULL,0,0,0),(1000,'Console',1,NULL,0,0,NULL),(1010,'NDS',1,NULL,0,0,1000),(1020,'PSP',1,NULL,0,0,1000),(1030,'Wii',1,NULL,0,0,1000),(1040,'Xbox',1,NULL,0,0,1000),(1050,'Xbox 360',1,NULL,0,0,1000),(1060,'WiiWare/VC',1,NULL,0,0,1000),(1070,'XBOX 360 DLC',1,NULL,0,0,1000),(1080,'PS3',1,NULL,0,0,1000),(1090,'Other',1,NULL,0,0,1000),(1110,'3DS',1,NULL,0,0,1000),(1120,'PS Vita',1,NULL,0,0,1000),(1130,'WiiU',1,NULL,0,0,1000),(1140,'Xbox One',1,NULL,0,0,1000),(1180,'PS4',1,NULL,0,0,1000),(2000,'Movies',1,NULL,0,0,NULL),(2010,'Foreign',1,NULL,0,0,2000),(2020,'Other',1,NULL,0,0,2000),(2030,'SD',1,NULL,0,0,2000),(2040,'HD',1,NULL,0,0,2000),(2050,'3D',1,NULL,0,0,2000),(2060,'BluRay',1,NULL,0,0,2000),(2070,'DVD',1,NULL,0,0,2000),(2080,'WEBDL',1,NULL,0,0,2000),(3000,'Audio',1,NULL,0,0,NULL),(3010,'MP3',1,NULL,0,0,3000),(3020,'Video',1,NULL,0,0,3000),(3030,'Audiobook',1,NULL,0,0,3000),(3040,'Lossless',1,NULL,0,0,3000),(3050,'Other',1,NULL,0,0,3000),(3060,'Foreign',1,NULL,0,0,3000),(4000,'PC',1,NULL,0,0,NULL),(4010,'0day',1,NULL,0,0,4000),(4020,'ISO',1,NULL,0,0,4000),(4030,'Mac',1,NULL,0,0,4000),(4040,'Phone-Other',1,NULL,0,0,4000),(4050,'Games',1,NULL,0,0,4000),(4060,'Phone-IOS',1,NULL,0,0,4000),(4070,'Phone-Android',1,NULL,0,0,4000),(5000,'TV',1,NULL,0,0,NULL),(5010,'WEB-DL',1,NULL,0,0,5000),(5020,'Foreign',1,NULL,0,0,5000),(5030,'SD',1,NULL,0,0,5000),(5040,'HD',1,NULL,0,0,5000),(5050,'Other',1,NULL,0,0,5000),(5060,'Sport',1,NULL,0,0,5000),(5070,'Anime',1,NULL,0,0,5000),(5080,'Documentary',1,NULL,0,0,5000),(6000,'XXX',1,NULL,0,0,NULL),(6010,'DVD',1,NULL,0,0,6000),(6020,'WMV',1,NULL,0,0,6000),(6030,'XviD',1,NULL,0,0,6000),(6040,'x264',1,NULL,0,0,6000),(6050,'Other',1,NULL,0,0,6000),(6060,'Imageset',1,NULL,0,0,6000),(6070,'Packs',1,NULL,0,0,6000),(6080,'SD',1,NULL,0,0,6000),(6090,'WEBDL',1,NULL,0,0,6000),(7000,'Books',1,NULL,0,0,NULL),(7010,'Magazines',1,NULL,0,0,7000),(7020,'Ebook',1,NULL,0,0,7000),(7030,'Comics',1,NULL,0,0,7000),(7040,'Technical',1,NULL,0,0,7000),(7060,'Foreign',1,NULL,0,0,7000),(7999,'Other',1,NULL,0,0,7000);
