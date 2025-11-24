
<map>
  <node ID="root" TEXT="Ch04-数据编码与演化">
    <node TEXT="问题背景" ID="94025d9b527ef8bba85877adf424c35a" STYLE="bubble" POSITION="right">
      <node TEXT="可演化性：当我们的程序功能更改时，存储的数据也需要更改" ID="5388b596900d21b1e80a064b6ac921ea" STYLE="fork"/>
      <node TEXT="序列化/反序列化：内存数据结构和写入到磁盘的数据结构不同，因此需要两者之间进行转化" ID="274f006b9ef5c7428f9c99253f414eb5" STYLE="fork"/>
    </node>
    <node TEXT="数据编码格式" ID="78df29e844285cf6865a3568d51486bc" STYLE="bubble" POSITION="right">
      <node TEXT="序列化与反序列化" ID="5ebf878ee060903bdd223d84d7662ad7" STYLE="fork">
        <node TEXT="在内存中，数据通常表示为结构体、对象等数据结构中" ID="b4ed09acbbf787c391ef2e266b52de73" STYLE="fork"/>
        <node TEXT="将数据写入文件或者通过网络发送时，必须将其编码为某种自包含的字节序列" ID="b6104e9424ea83d10f1b1b70f3c3ea7f" STYLE="fork"/>
        <node TEXT="序列化/编码：将数据从内存中的表示转化为字节序列" ID="d74cffbfc1073b638690cf9587433182" STYLE="fork"/>
        <node TEXT="反序列化/解码：将字节序列转化为内存中的数据结构" ID="ca571965e21982247f40e6b2bcade8bd" STYLE="fork"/>
      </node>
      <node TEXT="语言特定的格式" ID="c7074bde70b557ef12a006b2541373a9" STYLE="fork">
        <node TEXT="很多编程语言都支持序列化/反序列化操作" ID="32473a20227041570c1d90f4ea853c40" STYLE="fork"/>
        <node TEXT="存在的问题" ID="020677d45e7b0e4feaa9978aacce6688" STYLE="fork">
          <node TEXT="跨语言访问难" ID="a835f5cac369b5a9e16aa11e988aaff1" STYLE="fork"/>
          <node TEXT="容易导致安全问题" ID="2c35df4d9e364f9cecd391d247b5d4c2" STYLE="fork"/>
          <node TEXT="会有版本兼容问题" ID="fdf7523af7e7aeaff5a4e92fe8161168" STYLE="fork"/>
          <node TEXT="效率低" ID="d22831f36da81edde8e8083c7055b0c2" STYLE="fork"/>
        </node>
      </node>
      <node TEXT="JSON, XML和二进制变体" ID="e884308b1b9df4110b7baf0b391a498a" STYLE="fork">
        <node TEXT="JSON" ID="d740583b655e9009ff440a7531c87344" STYLE="fork">
          <node TEXT="语法" ID="04fdb6fd49823ad65e9e9b6d50e2b301" STYLE="fork">
            <node TEXT="数据以key-value形式组织" ID="bdf92befba8c9ef4d45f29293082af85" STYLE="fork"/>
            <node TEXT="对象使用{}包裹，数组用[]包裹" ID="8f58b72e56481eb76b2dd65303a477c8" STYLE="fork"/>
            <node TEXT="不支持注释" ID="f0ceb644b8816b57be301eb825fc9ad0" STYLE="fork"/>
            <node TEXT="不支持日期类型" ID="7b53a81508a639d0098d21c6b7680f9a" STYLE="fork"/>
          </node>
          <node TEXT="优点" ID="774973b890093d3917d47eecfc8be5a4" STYLE="fork">
            <node TEXT="简洁易读" ID="9c05e9864304d1c0d4677ecf890c76e5" STYLE="fork"/>
            <node TEXT="语言无关，支持广泛" ID="65f8cd52aa48b6406226f0eaf800664e" STYLE="fork"/>
            <node TEXT="数据结构灵活" ID="a91bab449624cea9629b19ec26bafc0d" STYLE="fork"/>
            <node TEXT="web友好" ID="6d7db4e065132083d291e3a6d8c5f137" STYLE="fork"/>
          </node>
          <node TEXT="缺点" ID="110bef01ed58b981dcda3bfca8da51b9" STYLE="fork">
            <node TEXT="数字精度问题" ID="9ae95e0b91cae26d76db1a5306c0e657" STYLE="fork"/>
            <node TEXT="没有日期类型" ID="12d3ab6a26f37eb4bd0c02e1148e7abf" STYLE="fork"/>
            <node TEXT="二进制数据支持差" ID="826af28d4a0e92572aed73648d2ca813" STYLE="fork"/>
            <node TEXT="没有注释" ID="316323dabebd52319ee935d20af33ab7" STYLE="fork"/>
            <node TEXT="没有引用和循环检测" ID="cd37215ef3be632e90efbccd3ab403df" STYLE="fork"/>
          </node>
          <node TEXT="适用场景" ID="74edc9736d339767ee40afbc9b95fdc7" STYLE="fork">
            <node TEXT="web api响应" ID="b8ee3d11d931a472652d54a13421b105" STYLE="fork"/>
            <node TEXT="配置文件" ID="9b0f219fc27b1da8245276e8257e1847" STYLE="fork"/>
            <node TEXT="NoSQL数据库" ID="a06a9b5cc1202838129ccb5b1839d687" STYLE="fork"/>
            <node TEXT="日志格式" ID="50b9067c7553b7a1e978d9604fe0b7a6" STYLE="fork"/>
          </node>
          <node TEXT="不适合" ID="68f5d94dcc14992a07d081264129d743" STYLE="fork">
            <node TEXT="大规模数据存储" ID="6068388dc4fc6aaf521a31c02afc27bf" STYLE="fork"/>
            <node TEXT="高性能RPC" ID="b469e2a7d4252965835dd746220fd0b8" STYLE="fork"/>
            <node TEXT="二进制数据传输" ID="a875ee525521a1a92394225361b176a3" STYLE="fork"/>
            <node TEXT="需要严格类型检查" ID="0f3a7c8dabeb1d8a378b1b5884c10268" STYLE="fork"/>
          </node>
        </node>
        <node TEXT="XML" ID="214c74d0e85a4fbf50e1019d6bc3450e" STYLE="fork">
          <node TEXT="优点" ID="da514a26a3858493fd4c7c6eba8f031b" STYLE="fork">
            <node TEXT="自描述性强" ID="8e3fa8510b3f8f9398b49f477d796a58" STYLE="fork"/>
            <node TEXT="支持复杂的数据结构" ID="eb4e5d57e7e43bed0c59003b3bbe1f76" STYLE="fork"/>
            <node TEXT="支持命名空间" ID="56914344cabc4eb8cb6ab59a8b17e063" STYLE="fork"/>
            <node TEXT="支持注释" ID="458ff7d45cd7ae4e2161b7bb9c04011d" STYLE="fork"/>
            <node TEXT="元数据丰富" ID="fe5c71e34d80b406d0f64bc2cf1cb126" STYLE="fork"/>
          </node>
          <node TEXT="缺点" ID="d9a463c9983706f920518f890ca1a004" STYLE="fork">
            <node TEXT="冗长，体积大" ID="f722a5f98da1bbce28aa314094da4f1c" STYLE="fork"/>
            <node TEXT="解析复杂" ID="7be97138ee5f7caad9df773a904418ec" STYLE="fork"/>
            <node TEXT="元素vs属性的模糊性" ID="e62feffd027140b7713425e661481a49" STYLE="fork"/>
            <node TEXT="数据类型弱" ID="ca93b0a947fcc10ca5648036c9858a3c" STYLE="fork"/>
            <node TEXT="没有数组的概念" ID="29e54c190b5f322d9206744657c6b5e3" STYLE="fork"/>
            <node TEXT="工具链复杂" ID="5ccb5d9c984f195a2877d0712f035a12" STYLE="fork"/>
          </node>
          <node TEXT="使用场景" ID="7470df1837aee0948c1190d70034895d" STYLE="fork">
            <node TEXT="配置文件" ID="a05281816e4e5df9dcd07726744dcd42" STYLE="fork"/>
            <node TEXT="文档格式" ID="96769ad19cd35fa803b2026edd35f2b4" STYLE="fork"/>
            <node TEXT="soap web services" ID="caa567fb1f53f4957e0445d4a3f2203e" STYLE="fork"/>
            <node TEXT="rss/atom订阅" ID="d5d0fb8112cf304e351bf546caa4ae85" STYLE="fork"/>
          </node>
          <node TEXT="不适用" ID="14f621df76eb55c59b7b6013d99c895b" STYLE="fork">
            <node TEXT="web api" ID="522721538a2ccfac075d31e1d8ace84e" STYLE="fork"/>
            <node TEXT="高性能场景" ID="a1677122b0028a21acd6e5d775a2da70" STYLE="fork"/>
            <node TEXT="简单数据结构" ID="8f94bfc76180b04e670485e08f0c5cee" STYLE="fork"/>
            <node TEXT="现代微服务" ID="eea4294fa10dfb01cc59240f1ccd42c8" STYLE="fork"/>
          </node>
        </node>
        <node TEXT="二进制编码" ID="19c3fd7bde0c944c0173f85f0098ef8e" STYLE="fork">
          <node TEXT="二进制编码的优势" ID="482a6fba888622a63d71fdced3a310cf" STYLE="fork">
            <node TEXT="体积小" ID="7dab3ec38504c146285e6c300ea637f8" STYLE="fork"/>
            <node TEXT="类型强" ID="7a24d2e6ecee5ee3478914c4dcafe790" STYLE="fork"/>
            <node TEXT="解析快" ID="7c37a36dbb036253eff4d523ed80ff78" STYLE="fork"/>
            <node TEXT="向前/向后兼容" ID="137e168c03676c0d107b61b5810d01fe" STYLE="fork"/>
          </node>
          <node TEXT="Protocol Buffers" ID="00b673c2e43faab31014d6a3b50dfc27" STYLE="fork">
            <node TEXT="编码格式" ID="697c2258562a328ead8ad8f2665ca59e" STYLE="fork">
              <node TEXT="[Tag] [Value]" ID="801b14854d41c8dfaa6934cad082e8ef" STYLE="fork"/>
              <node TEXT="Tag = (field_number &lt;&lt; 3) | wire_type" ID="a757efcc1cc4ba671ff1ff926a4b8150" STYLE="fork"/>
            </node>
            <node TEXT="缺点" ID="9f6616ab5532054f07896b9b744d8a51" STYLE="fork">
              <node TEXT="不可读" ID="77cacf3600bfd37172f5902bfabec3d2" STYLE="fork"/>
              <node TEXT="需要schema" ID="d0f1bbdb74387ad4fdec11872718dbf0" STYLE="fork"/>
              <node TEXT="不适合大字段" ID="86efcc5609385f02c935a08998b933ab" STYLE="fork"/>
            </node>
            <node TEXT="适用场景" ID="e23104248abfcdefa0f2b3ab89e98d1d" STYLE="fork">
              <node TEXT="gRPC" ID="e7a03d321b720aa907557ef72c57560b" STYLE="fork"/>
              <node TEXT="消息队列" ID="e26ae989adbac110ab694ed9b822d719" STYLE="fork"/>
              <node TEXT="数据库存储" ID="3481b3b415add8c4465e8d770f3828c2" STYLE="fork"/>
              <node TEXT="移动应用api" ID="b85a3d83833a1c6a8acd2c6c9b976a9b" STYLE="fork"/>
            </node>
          </node>
          <node TEXT="Thrift" ID="2f2a86c5d736a24632574551ddbf5419" STYLE="fork">
            <node TEXT="编码格式" ID="040e039060b063da3ca2361e49d8e3cc" STYLE="fork">
              <node TEXT="[字段类型] [字段ID] [字段值]" ID="e4885fc03b4ae4feb3acf78b4fbdf524" STYLE="fork"/>
            </node>
            <node TEXT="缺点" ID="1d7fff0bfdc3f34eb7ff7c9a3c83d47f" STYLE="fork">
              <node TEXT="生态系统较小" ID="6b42a7c0cd422b3e9310dd238313b2eb" STYLE="fork"/>
              <node TEXT="required 字段陷阱" ID="ce1400b7991e54470c9ecaf4626b5780" STYLE="fork"/>
              <node TEXT="文档和工具不如 protobuf" ID="24deb2b4b5413a0679b572668c55b565" STYLE="fork"/>
            </node>
          </node>
        </node>
        <node TEXT="模式(Schema)的优点" ID="51f5309fd2ec623d5098c26500f04f70" STYLE="fork">
          <node TEXT="schema是对数据结构的明确约定，它保证了编译时的类型安全、演化时的兼容性检查以及运行时的高效编码" ID="a8c109a1796f83430b0af9008cd580ec" STYLE="fork"/>
          <node TEXT="优点" ID="396626f03e376f74172b6cc72d823c85" STYLE="fork">
            <node TEXT="更紧凑的编码" ID="5dc9936bbe181617dbaa2b04cfdda1e0" STYLE="fork">
              <node TEXT="对比json/xml，schema使用字段编码代替字段名，节省了大部分空间" ID="33dd3d9723e84a449b81931f712694ea" STYLE="fork"/>
            </node>
            <node TEXT="schema代码即文档" ID="9ab4a34324e01bbc93330e602e108c8d" STYLE="fork">
              <node TEXT="" ID="926a2fa6de2425724451378904b38709" STYLE="fork"/>
            </node>
            <node TEXT="schema是代码生成源头，保持数据库的数据一直是最新，数据结构变化代码也要重新生成" ID="896e2c28c0e3a160ca78a63d2ed6de98" STYLE="fork"/>
            <node TEXT="允许前向/后向兼容性检查" ID="fec6ed544c1d9bc58c538d825dc0469f" STYLE="fork">
              <node TEXT="使用json，一旦新版本悄悄改了字段名称则可能会导致线上事故" ID="98eb129a1ac17ed768339c4e61b986fd" STYLE="fork"/>
            </node>
            <node TEXT="自动生成代码，避免重复的样板代码" ID="8beda81833114c72e1ff30b05eb29152" STYLE="fork"/>
          </node>
          <node TEXT="何时需要schema" ID="fd4347141dac7be8e8eec269a3480efe" STYLE="fork">
            <node TEXT="微服务通信" ID="932ee276097cfec22c1ad48273bff5dc" STYLE="fork"/>
            <node TEXT="高性能要求" ID="beaebe56a2dbe1e9caafb8db134aaa9e" STYLE="fork"/>
            <node TEXT="团队协作" ID="f85dd986a718bc5ec507a1fd552c044f" STYLE="fork"/>
            <node TEXT="长期维护的系统" ID="bf27817841a4771d2a2e889e0f3e48d9" STYLE="fork"/>
            <node TEXT="需要演化的API" ID="022e7941da2eaa8d1c92b1d371e56f5c" STYLE="fork"/>
          </node>
        </node>
      </node>
    </node>
    <node TEXT="数据流模式" ID="faa05b2d5f3432a6c0ba58818ee20344" STYLE="bubble" POSITION="right">
      <node TEXT="现实挑战" ID="1f6b5447137cae8647f3fb6a00b68825" STYLE="fork">
        <node TEXT="两个服务之间如何传递数据？如何保持兼容" ID="691be2c4d116d84919b96a902218f1ae" STYLE="fork"/>
        <node TEXT="服务版本可能不同" ID="a16b47c74e419d216d24f422eb6d11c4" STYLE="fork"/>
        <node TEXT="服务可能在不同机器上" ID="8e54a32813a9f4351401920b353e3906" STYLE="fork"/>
        <node TEXT="服务可能使用不同语言" ID="84986aef699e1713fecca7bc1ac13f72" STYLE="fork"/>
        <node TEXT="不能同时停机升级" ID="9533c738d2d74effd7c38af16675c851" STYLE="fork"/>
      </node>
      <node TEXT="进程间数据流动方式" ID="1fa222d448acb9db63084bb1f1199fc5" STYLE="fork">
        <node TEXT="通过数据库" ID="95eae922d7d400a059fb59424f00b8e7" STYLE="fork"/>
        <node TEXT="通过服务调用" ID="f3e3125087373c9f503f8df0d65e433f" STYLE="fork"/>
        <node TEXT="通过异步消息传递" ID="c12036b96467a2200840021596d30e5a" STYLE="fork"/>
        <node TEXT="" ID="09c514b2dcd9c5b2bd1edfa22d86c641" STYLE="fork"/>
      </node>
      <node TEXT="数据库的数据流" ID="4f60a0156d51db3b31e8c7d9acdc6c0e" STYLE="fork">
        <node TEXT="不同的应用程序通过读写数据库实现数据流动" ID="c4eabad5bc0e9921ebd178bc6a50fa63" STYLE="fork"/>
        <node TEXT="核心问题" ID="0353df460ac064b35d36d5d17d33a8d1" STYLE="fork">
          <node TEXT="未来的自己读取过去的自己写的数据" ID="0700eccf309240440bac62aafb65b82e" STYLE="fork">
            <node TEXT="未来程序增删了数据段怎么办？" ID="dd37c9a2f1fa4aeca413d36d576739bc" STYLE="fork"/>
          </node>
          <node TEXT="新旧代码并存" ID="644d33ef94e17ae62dd3ac62f684eb1b" STYLE="fork"/>
        </node>
        <node TEXT="解决方案" ID="28f91b3087f793a84e323a8d47ba69fa" STYLE="fork">
          <node TEXT="向前/向后兼容" ID="021d9b77b563ceae0f9bcadd54060652" STYLE="fork"/>
          <node TEXT="尽量避免重写数据库" ID="8c15c7050c4f9dc445423f6d58cb2607" STYLE="fork"/>
        </node>
      </node>
      <node TEXT="服务调用的数据流" ID="5950f718384f6340c3c125063c3ccb76" STYLE="fork">
        <node TEXT="服务器通过网络公开API，客户端可以连接到服务器以向该API发出请求" ID="333822c04e009e2bbdab343203f41714" STYLE="fork"/>
        <node TEXT="核心问题" ID="7286d1624a150c715d546574ea0317a0" STYLE="fork">
          <node TEXT="客户端和服务端版本不同，保证能够工作" ID="5592d93c8438f45feecbaf9ac4a86dc0" STYLE="fork"/>
          <node TEXT="服务依赖链" ID="c3b2508712bf927d08e97744c244d546" STYLE="fork"/>
        </node>
        <node TEXT="REST" ID="5f8c16d2557f16d4aa1801daa59a9d11" STYLE="fork">
          <node TEXT="特点" ID="c87bb938be3e00ddac8a0521f324a0ad" STYLE="fork">
            <node TEXT="基于HTTP" ID="03208b626cd0863322d9cc0c087c3971" STYLE="fork"/>
            <node TEXT="使用JSON/XML" ID="5816488a1f4ffef6df2bc8a4a697e442" STYLE="fork"/>
            <node TEXT="自描述性强" ID="a8a0a7df2840f5fbf56f7ddd33604c54" STYLE="fork"/>
            <node TEXT="浏览器友好" ID="a9a5372f68a72602abd4323ca5d5963c" STYLE="fork"/>
          </node>
          <node TEXT="兼容性" ID="6b05e3570ed61dd4c600ff1a2627d1b1" STYLE="fork">
            <node TEXT="向后兼容：新增字段不影响旧客户端" ID="9936a17253e63e85ef4101292ea00f68" STYLE="fork"/>
            <node TEXT="向前兼容：旧客户端忽略新字段" ID="61d1f687199d79eaf5115b24543881a8" STYLE="fork"/>
            <node TEXT="问题：无强制schema，容易出错" ID="0cd148c38059f760fd8b9cbca394c016" STYLE="fork"/>
          </node>
        </node>
        <node TEXT="RPC" ID="28dff1fbeef9f289fcba01480c5cffa7" STYLE="fork">
          <node TEXT="特点" ID="fe787f882c8cfb5fd7badb12f9799a79" STYLE="fork">
            <node TEXT="调用远程函数就像本地函数" ID="24bd8985bceb95fca93ca68eb2d83082" STYLE="fork">
              <node TEXT="与本地函数的不同" ID="e93dfbdfe0fedae4312e4a6480f3d328" STYLE="fork">
                <node TEXT="网络请求不可预测，可能导致请求失败；而本地函数的结果是可预测的" ID="3a8a3deeb72da8f5639d66d7c2a5e6de" STYLE="fork"/>
                <node TEXT="由于网络超时的存在，我们可能永远无法知道RPC调用发生了什么（是没有收到请求/调用失败/网络超时）" ID="7e7697f761b44b3dc39bbe8b7c4b8dbb" STYLE="fork"/>
                <node TEXT="每次调用的时间不可预测" ID="3ea529a9c3890f5580206c8528c24995" STYLE="fork"/>
                <node TEXT="本地函数可以使用指针来传递内存中的大对象；RPC必须经过序列化才能在网络中传输对象" ID="df7c411cc5646243968890192221f13c" STYLE="fork"/>
              </node>
            </node>
            <node TEXT="通常使用二进制协议" ID="201a856062fb60d325a95f538ff05203" STYLE="fork"/>
            <node TEXT="性能好" ID="6ead7627856dd3cf6ba27a31a7972fd3" STYLE="fork"/>
            <node TEXT="需要schema" ID="151466a6c3e89b4ebd7340c5888579b3" STYLE="fork"/>
          </node>
          <node TEXT="兼容性" ID="3f5e35e53f29edca2af3d0f7da9da583" STYLE="fork">
            <node TEXT="强类型schema保证兼容性" ID="c28ccfb012a3954cd4382d474182fb7a" STYLE="fork"/>
            <node TEXT="编译时检查" ID="6cec7125fb4009a57e1375360582193a" STYLE="fork"/>
            <node TEXT="问题：需要客户端和服务端都升级代码" ID="2a4bcbea455390f54f3025e34e74eb93" STYLE="fork"/>
          </node>
          <node TEXT="解决方案" ID="89b29b205dc11cf1a304977b51a586d7" STYLE="fork">
            <node TEXT="API版本管理" ID="d2b37e8d29b5751052284516452d6abb" STYLE="fork">
              <node TEXT="url版本化" ID="9dac10a71921e63df22d3b588b4cf167" STYLE="fork"/>
              <node TEXT="兼容性演化" ID="e052f870127bd1b7a9ce41d3b1c34a73" STYLE="fork"/>
              <node TEXT="GraphQL" ID="7256ce9e72cdf9f0efae0a5917276410" STYLE="fork"/>
            </node>
          </node>
        </node>
      </node>
      <node TEXT="消息传递的数据流" ID="13b113d2de8e21f39402adf3dc7a8330" STYLE="fork">
        <node TEXT="数据不是通过直接的网络连接发送，而是通过被称为消息代理的中介发送，中介会暂存消息" ID="cfdcfc59706b2768bb94504a532f296f" STYLE="fork"/>
        <node TEXT="核心问题" ID="7f16ec5f5aa1fa51cc8703e683796237" STYLE="fork">
          <node TEXT="生产者和消费者版本不同" ID="a0ebce972f4f8cb56c7683912c70f463" STYLE="fork"/>
          <node TEXT="消息积压和重放" ID="101f1918a33c4f7ef1f81359a6e67bac" STYLE="fork">
            <node TEXT="消费者宕机，消息在队列中积压" ID="0b69fd7dd9891d1375963e7d2bd14d2c" STYLE="fork"/>
            <node TEXT="积压的消息可能是旧版本" ID="e614df1ee4796c72464ab480650abd7e" STYLE="fork"/>
          </node>
        </node>
        <node TEXT="解决方案" ID="70bc40a98ac2240373aff85202b80591" STYLE="fork">
          <node TEXT="消息格式演化" ID="3cd40d3b69237c0e4acb957651504d1d" STYLE="fork">
            <node TEXT="Avro（动态Schema）" ID="159f46662455f56ee4fd2719ff23cbe2" STYLE="fork"/>
            <node TEXT="Schema Registry" ID="a755f4f9c1fb3899fedd046f76fb412c" STYLE="fork">
              <node TEXT="生产者发送消息前，先注册schema" ID="428aa0d67a629108e468006db779c272" STYLE="fork"/>
              <node TEXT="消息头包含schema ID" ID="59501a9d7651259078b6b7064e7349aa" STYLE="fork"/>
              <node TEXT="消费者根据schema ID读取schema" ID="3ba5f596946af94bfbaf657936ff20aa" STYLE="fork"/>
              <node TEXT="schema registry检查兼容性" ID="971785f5b7ef9ab08dc018c9d8406898" STYLE="fork"/>
            </node>
          </node>
        </node>
      </node>
      <node TEXT="核心设计原则" ID="b7ed5e412937177fb9a50cd95a8d95f2" STYLE="fork">
        <node TEXT="永远假设版本不同" ID="7d65de310292aad7322f7628c7dea698" STYLE="fork"/>
        <node TEXT="向后兼容优先" ID="a0105ef465df9b07f23b951eab351f8a" STYLE="fork">
          <node TEXT="新代码读旧数据是常态" ID="a7e4d8006a4819ec3661aced57b8a892" STYLE="fork"/>
          <node TEXT="历史数据迁移成本高" ID="7e4ee2f6e36f1dc0a731735994463a9d" STYLE="fork"/>
        </node>
        <node TEXT="schema演化规则" ID="fd764940c72c35c302a48e99d48693d8" STYLE="fork">
          <node TEXT="允许" ID="13d1b92f01375a1e33a15250152ae234" STYLE="fork">
            <node TEXT="新增可选字段" ID="ad7ddb108b16f21981fea9aed207f13c" STYLE="fork"/>
            <node TEXT="删除可选字段" ID="1a11acb55967bb121dd0057a7eeaceac" STYLE="fork"/>
            <node TEXT="修改字段名（保持ID/编号不变）" ID="9ee19f5e397830a39bed2434853cdb06" STYLE="fork"/>
          </node>
          <node TEXT="禁止" ID="2b6c5c124ac6cab5b848124119764dfa" STYLE="fork">
            <node TEXT="修改字段类型（不兼容）" ID="4694b7e70fc3fc909787c66f625d9ccd" STYLE="fork"/>
            <node TEXT="删除必须字段" ID="ecfe795487e2c049244b681dd46a3750" STYLE="fork"/>
            <node TEXT="修改字段编号/ID" ID="2ae8a6eb9ffb075c8c8ff54f6d1d7613" STYLE="fork"/>
            <node TEXT="改变字段语义" ID="8ea9a63d25ddfd1fd8219e886a63d22c" STYLE="fork"/>
          </node>
        </node>
        <node TEXT="包含schema版本信息" ID="8a4b0fccb9369f85b874d968af145e8a" STYLE="fork"/>
      </node>
    </node>
  </node>
</map>