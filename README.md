# pubsrc

pubsrc(public services) 为foxone提供一些基础的公开服务

## Price

- Use `wallet_2019-10-24.sql` to create table `assets`
- Fill the column `cmc_slug` with command `slug`
- Use the commands in `extra.sql` to add missing slugs
- Run the command `price` to get history price and save them to table `*_snapshots`
- Run the command `service` to get lastest price 

