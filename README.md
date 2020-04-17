# trickydns
A tricky DNS server runs only in local to anti DNS response cheat from GFW.

利用了GFW对DNS压缩指针的不严谨处理，来让GFW认为经过改造的请求并非一个合法的DNS请求从而不会伪造DNS响应，从而不需要服务器即可访问一些被屏蔽的网站。效果类似于改hosts，好处是不需要手动维护列表。而且由于DNS请求的源IP为真实IP，所以CDN也能正常工作。

由于不需要判断获得的DNS响应是否是伪造的，所以比通过DNS响应的先后顺序来丢弃伪造的响应更稳定一些。

解决了DNS污染后能直接访问的网站有github gist、feedly等，加上 https://github.com/bypass-GFW-SNI/proxy 后能直接访问的网站有wikipedia、pixiv等。欢迎在Issues里补充。

技术细节：

经过测试，现在GFW最多只能处理17个级联指针，超过17个指针后GFW便不再对此DNS请求再进行处理，可能是为了避免循环指针导致死循环或消耗太多资源。trickydns通过改造DNS请求，将原始域名移动位置，再通过18个以上的级联指针指向原始域名，从而避免了GFW发送伪造的DNS响应。

使用方法，以Linux为例：

`# nohup ./trickydns trickydns.config.json &`

即可在本机开启一个无污染的DNS服务。

trickydns.config.json文件的配置例子可以查看： https://github.com/lehui99/trickydns/blob/master/trickydns.config.json
