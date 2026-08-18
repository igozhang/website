# Post_Views_ExportImport

wordpress 插件Post Views Counter 浏览数据导入导出  
wordpress 调试

# 导出 Export

```
mysqldump -uroot -ppassword wordpress wp_post_views > /tmp/ali_views.sql;
-- 导出wordpress数据库wp_post_views表的数据到ali_views.sql

sz下载
```

# 导入 Import

```
rz上传

--登录mysql进入wordpress数据库
mysql -uroot -ppassword wordpress
--备份原数据
mysqldump -uroot -ppassword wordpress wp_post_views > /tmp/qq_views.sql;
--导入数据
source /root/ali_views.sql

--或者直接在shell导入
mysql -uroot -ppassword wordpress < /路径/ali_views.sql
```

## 修改底部主题信息

1. 隐藏底部主题信息  
   如果主题支持自定义，找到自定义footer部分，设置就行  
   如果主题不支持自定义，找到主题对应的footer.php文件位置，已spicepress为例

```
位置为：/usr/local/lighthouse/softwares/wordpress/wp-content/themes/spicepress/footer.php

关键行
$spicepress_footer_copyright = get_theme_mod('footer_copyright_text','<p>'.__( 'Proudly powered by <a href="https://wordpress.org">WordPress</a> | Theme: <a href="https://spicethemes.com" rel="nofollow">SpicePress</a> by SpiceThemes', 'spicepress' ).'</p>');
替换后
$spicepress_footer_copyright = get_theme_mod('footer_copyright_text','<p>'.__( 'igozhang 2011' ).'</p>');

改完无需重启
```

## 自定义头像

1. 安装插件Simple Local Avatars 就可以支持自定义用户头像

## 自定义管理员邮箱

```
-> mysql -uroot -p
-> use wordpress
-> update wp_users set user_email=igozhang@163.com;
-> FLUSH   PRIVILEGES;
```

## 修改网站地址错误导致的跳转问题修复

```
手动修改回来
ssh root
wp option update siteurl http://150.109.244.111 --allow-root
wp option update home http://150.109.244.111 --allow-root
```

## HTTPS证书问题

1. 需要先在宝塔面板打开SSL，才会监听443端口
2. 然后使用命令修改网站及后台地址

```
wp option update siteurl https://igozhang.cn --allow-root
wp option update siteurl https://igozhang.cn --allow-root
```

## dls

```
文件位置从nginx.conf 的include部分去找
location /dls {
    alias   /data/dls/;
    autoindex on;
    autoindex_exact_size off;
    autoindex_localtime on;
    autoindex_format html; #以html风格将目录展示在浏览器中
    charset utf-8,gbk; #展示中文文件名
}

nginx -s reload
```
