# toml2php

一个简单的toml转php代码库，用于将toml转换为php代码，可用于PHP项目的配置系统等。

本项目参考了 https://github.com/leonelquinteros/php-toml 这个库，并基于这个库的实现进行了一些调整。与原库自动解析为PHP对象不同的是，toml2php的开发初衷是开发一个库，用于PHP项目的配置系统，这相当于需要一个第三方工具将toml配置抓换为PHP代码，所以在目的和实现上是有本质区别的。toml2php主要用于将toml配置转换成对应的php代码。

toml2php与上面的php-toml的不同：

1. 目的不同：toml2php用于辅助的配置转换，用于生成可执行的PHP代码，不是直接将toml转换成内部使用的php对象；

2. 支持的数据类型有所不同，toml暂未支持下面的类型：
    
    * 日期时间类型，如果需要，可以使用字符串，然后自行处理；
    
    * 由下划线格式化的数字；
    
    * ~~科学计数法表示的数字；~~

toml2php的使用者在使用时，必须明确指出解析的内容是单个值还是数组，并据此调用ParseSingle或ParseTable方法。


## [ChangeLog]

* 2020.01.31 添加科学计数法支持；
 
