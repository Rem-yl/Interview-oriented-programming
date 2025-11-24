
<map>
  <node ID="root" TEXT="Ch06-数据分区">
    <node TEXT="背景" ID="c53f45d0c79082b3b6c580c76a7507ce" STYLE="bubble" POSITION="right">
      <node TEXT="单机存储的局限性" ID="d66bb71aa7d511b6b55f327833d8f62c" STYLE="fork">
        <node TEXT="数据量超过单机磁盘容量" ID="45ca6756a2b8908d282fde2347ade69c" STYLE="fork"/>
        <node TEXT="查询吞吐量超过单机处理能力" ID="dc72dc89e65e5c25414c9b0eb97d09b6" STYLE="fork"/>
        <node TEXT="无法水平扩展" ID="ffca38f76b144ee938881a365e031db5" STYLE="fork"/>
        <node TEXT="单点故障风险" ID="2fa91df125beef3032b202f7b4b98139" STYLE="fork"/>
      </node>
      <node TEXT="数据分区的核心思想" ID="776f951630302ab1bc358c2fc715077a" STYLE="fork">
        <node TEXT="分区=分片" ID="defab96ecbc47453988f6899a0569f6c" STYLE="fork"/>
        <node TEXT="数据分散存储" ID="9cdcda1770983d5eb2534604db44070b" STYLE="fork"/>
        <node TEXT="查询并行处理" ID="f7f490aa4afae7193127cbd1816410ca" STYLE="fork"/>
        <node TEXT="配合复制，提高可用性" ID="6fededd95a8fe9f5d2a4599be1be59de" STYLE="fork"/>
      </node>
      <node TEXT="分区的核心目标" ID="b8365515ffaa5a2309b4d3f86d9a2976" STYLE="fork">
        <node TEXT="可扩展性" ID="1a645ac863b00db65be0fc134893547b" STYLE="fork"/>
        <node TEXT="负载均衡" ID="17515f721e6e14ad464c1baaa65606d4" STYLE="fork">
          <node TEXT="每个节点存储的数据量相近" ID="469ae506c580a2c37ddb6a207a5feecf" STYLE="fork"/>
          <node TEXT="每个节点处理的请求量相近" ID="e0967e0a73ef0305b4599b7d4c8148d6" STYLE="fork"/>
        </node>
        <node TEXT="高可用性" ID="c808b85fef9c848eebebb6ff37cf1fc2" STYLE="fork"/>
      </node>
    </node>
    <node TEXT="分区策略" ID="f7c65989c7a7aec0d909028050416dcc" STYLE="bubble" POSITION="right">
      <node TEXT="按键范围分区" ID="4b8429faf9566f6575c695481716d78f" STYLE="fork">
        <node TEXT="将连续的键范围分配给不同的分区" ID="82a03b3c412d5c7732250a2dd04ceee6" STYLE="fork"/>
        <node TEXT="优点" ID="ee4de693fa6566bc3bc6b9ff39fe1bdf" STYLE="fork">
          <node TEXT="范围查询高效" ID="35206e6fe7e567fe114fdadf0adeb237" STYLE="fork"/>
          <node TEXT="数据在分区内有序" ID="c2f9d0997e211d7d4435420e02d2dc6d" STYLE="fork"/>
          <node TEXT="分区边界可以灵活调整" ID="bffd2ca4c36d03d52ca6273605330443" STYLE="fork"/>
        </node>
        <node TEXT="缺点" ID="667cd40d9bbc94d3353b4d89bda6a316" STYLE="fork">
          <node TEXT="热点问题" ID="203521c44fab95f80ecb3b567a3869dc" STYLE="fork">
            <node TEXT="某个名人的数据分区可能负载特别高" ID="a6dc3595b4ae7526082cb30d478f424b" STYLE="fork"/>
            <node TEXT="负载不均衡" ID="e23a1086862fd539da9944a7a525b3bd" STYLE="fork"/>
            <node TEXT="解决办法" ID="79180e118fd05142b6d7d3360feca7cc" STYLE="fork">
              <node TEXT="键前缀随机化" ID="1253bfd292946a9dcad0f828e6924d0b" STYLE="fork">
                <node TEXT="缺点：范围查询需要扫描所有分区" ID="bfca99e0073e981334eba530cbb3040e" STYLE="fork"/>
              </node>
              <node TEXT="符合键：使用两个键组合作为分区键" ID="149a58fa759b9a2e3cb0504be539c68d" STYLE="fork"/>
            </node>
          </node>
        </node>
      </node>
      <node TEXT="按键哈希值分区" ID="948ef33076408ce4a18d9ad6944ba0ca" STYLE="fork">
        <node TEXT="使用哈希函数将键均匀分配到分区" ID="b399410456c49dcc02e0dca45b82f943" STYLE="fork"/>
        <node TEXT="一致性哈希：解决普通哈希在节点变化时大量数据迁移的问题" ID="be144aac7e2c1cfc57937edb779045cd" STYLE="fork"/>
        <node TEXT="优点" ID="eadd679f8f27e3af031d07168aebb880" STYLE="fork">
          <node TEXT="数据分布更均匀" ID="cc304c7d0e848d1c8af2710eedb9ccd9" STYLE="fork"/>
          <node TEXT="节点故障时负载分散到多个节点" ID="14cda6b3e84e54d3a541ff39fd29609f" STYLE="fork"/>
          <node TEXT="负载均衡" ID="0748edc3f0de75521fd3550f1e5d0009" STYLE="fork"/>
        </node>
        <node TEXT="缺点" ID="43170ff26b98ab13f224da3e94ca1b98" STYLE="fork">
          <node TEXT="无法进行范围查询" ID="e843847d7b98cb798a1049c9138777b4" STYLE="fork"/>
          <node TEXT="相邻键分散存储" ID="499407f82c4543c6de3b526b0c8a86ee" STYLE="fork"/>
          <node TEXT="节点增减时数据迁移" ID="95d9baf55a112487edcf338a8a5a0b06" STYLE="fork"/>
        </node>
      </node>
      <node TEXT="混合策略：复合主键分区" ID="c10a7223360806bdde22b4ad34b2ef16" STYLE="fork">
        <node TEXT="结合哈希和范围分区的优点" ID="e3bee64cc8095817a823df0cef6b13be" STYLE="fork"/>
        <node TEXT="按照user_id进行哈希分区" ID="805aedb99fe25dead4d307f740942faa" STYLE="fork"/>
        <node TEXT="分区内按照timestamp排序" ID="b47034861aa5a0ccba6ca56c0c8bf3d3" STYLE="fork"/>
      </node>
      <node TEXT="按工作负载分区" ID="3e50d53cbfa43b5a3158324cfb4c0491" STYLE="fork">
        <node TEXT="场景：名人用户的热点问题" ID="59eebe6ac63fe96f10873a783f450410" STYLE="fork"/>
        <node TEXT="解决方案" ID="c2b95a4829840505e0ed733c915a28b4" STYLE="fork">
          <node TEXT="手动分区" ID="ca9141ab4ecd93e23f3b1417e8f0d500" STYLE="fork">
            <node TEXT="手动将名人用户放到不同的分区" ID="4c8c1ec2f5ea69591899a4b17508d073" STYLE="fork"/>
            <node TEXT="需要手动维护热点列表" ID="567525d991455e5b8d99c82cec043a65" STYLE="fork"/>
          </node>
          <node TEXT="数据复制" ID="df90eb764064e78ba8566338bc90f839" STYLE="fork">
            <node TEXT="将热点数据复制到多个分区" ID="5418128d9a1d6a4d51610c884dca9114" STYLE="fork"/>
            <node TEXT="更新复杂" ID="b42be6f2465a5602e0e9d8d3553f2b1b" STYLE="fork"/>
          </node>
          <node TEXT="应用层缓存" ID="a242b77cef96ddbe9bc45c9ec4ae19b7" STYLE="fork">
            <node TEXT="可以将热点数据缓存到redis" ID="a6085c9b890f33e3541c47378888f7cf" STYLE="fork"/>
          </node>
        </node>
      </node>
    </node>
    <node TEXT="分区与二级索引" ID="4685a432a0f91b9a730d40abcee2cad3" STYLE="bubble" POSITION="right">
      <node TEXT="问题场景：分区按照主键分区，使用二级索引查询时需要扫描所有分区" ID="6fad7e854ab28cdf1df326414b0d63c3" STYLE="fork"/>
      <node TEXT="解决方案" ID="c7d3f2dec5effdeb1d56335471700895" STYLE="fork">
        <node TEXT="每个分区维护自己的本地索引" ID="259a1647694ef3eea467e3471b737718" STYLE="fork">
          <node TEXT="流程" ID="40d1dccf15759d9d284fced208667e3a" STYLE="fork">
            <node TEXT="" ID="710ace2cdf001effbbc3a2744b69a146" STYLE="fork"/>
            <node TEXT="查询语言广播到所有的分区，合并查询返回的结果" ID="62eccd6a644340d07ed287838ad59a14" STYLE="fork"/>
          </node>
          <node TEXT="优缺点" ID="0483597f1bac696f37238bec727022a6" STYLE="fork">
            <node TEXT="写入简单；索引维护容易" ID="add9ab623993d2f30121aeff150efc14" STYLE="fork"/>
            <node TEXT="读取低效，需要查询所有的分区" ID="dd3c5bc243a7ef1f5b11275e89f91aed" STYLE="fork"/>
            <node TEXT="需要应用层去重" ID="a5fe0ae4298b378a0aafeefa81c34352" STYLE="fork"/>
          </node>
        </node>
        <node TEXT="全局分区索引" ID="5926e2bc254e8b69d0b93bd11419115e" STYLE="fork">
          <node TEXT="流程" ID="03a2c43220e8f81ec9b9b2a20ce18e18" STYLE="fork">
            <node TEXT="索引本身也进行分区" ID="2fa6fd62234dab4d3a39b3142289166e" STYLE="fork"/>
            <node TEXT="根据user_id查询主数据" ID="af05be60e665844d9195176d0a7791a9" STYLE="fork"/>
            <node TEXT="查询email时，先计算hash(email) = 523，落在索引分区A" ID="1a16a151bd931baddaed54379fd22f58" STYLE="fork"/>
            <node TEXT="查询索引分区A：返回user_ID[1,6]" ID="1d25f1c7693ce1d7053621f90cc63fd4" STYLE="fork"/>
            <node TEXT="优缺点" ID="fb2e5c15b85f7d5467e13ccd6be99ea0" STYLE="fork">
              <node TEXT="读取高效：只需要查询一个索引分区" ID="452f8481c62d5797825b2d2269d146dc" STYLE="fork"/>
              <node TEXT="负载均衡：索引查询分散到多个分区" ID="92bfffa31366c559cc6ef6dd179c01ca" STYLE="fork"/>
              <node TEXT="写入复杂：需要跨分区事务" ID="c18b3f7a6a2760b6f77b8f80f089c6c5" STYLE="fork"/>
            </node>
          </node>
        </node>
        <node TEXT="分区索引的均衡" ID="42ec73fd53cea13c78bf32cf590fd11c" STYLE="fork">
          <node TEXT="" ID="7546707a759033ca8b59f0cdc99bb9bf" STYLE="fork"/>
        </node>
      </node>
    </node>
    <node TEXT="分区再平衡" ID="d25fe43f1742b1d0b9d4086ac92bca37" STYLE="bubble" POSITION="right">
      <node TEXT="背景" ID="1de8ac17c9bc6a8a807eef1225385cdb" STYLE="fork">
        <node TEXT="数据增长需要增加节点" ID="a76fb05c8459116f076aea069baacee6" STYLE="fork"/>
        <node TEXT="节点故障" ID="49fcab7cf724b2cdc95c441b393ddc05" STYLE="fork"/>
        <node TEXT="负载不均衡时，需要调整分区" ID="b8b312f3c9eb782374aa210f12b13661" STYLE="fork"/>
      </node>
      <node TEXT="要求" ID="fafdaea188da52adb18b6795a5a76088" STYLE="fork">
        <node TEXT="负载均衡：再平衡后每个节点的负载相近" ID="84738cedab33373251fbf3a37ae77538" STYLE="fork"/>
        <node TEXT="最小化数据迁移" ID="5f1fb0c5079cf4e2262f7c970e83c62b" STYLE="fork"/>
        <node TEXT="系统可用：再平衡期间系统不能停机" ID="6e9d4a94182612854c5d9979c69409f8" STYLE="fork"/>
        <node TEXT="自动化" ID="092352b30458144b6b66fd8e776606cf" STYLE="fork"/>
      </node>
      <node TEXT="再平衡策略" ID="c95bc220487d7adb7b19d600689566d4" STYLE="fork">
        <node TEXT="固定数量的分区" ID="4b6470da7d8cfc7a05af9688c4a295ad" STYLE="fork">
          <node TEXT="预先创建大量分区，节点增减时移动分区" ID="e4a05aadb83debb2c0c271f498a20e18" STYLE="fork"/>
          <node TEXT="优点" ID="c5ea6dc124b6a356ce3ff13a2ad76d4d" STYLE="fork">
            <node TEXT="简单；分区数固定，只需要移动整个分区" ID="e0e4ac9634dd8bcae76497c6ee76dea3" STYLE="fork"/>
            <node TEXT="迁移量可控" ID="8f4c56bdd27294767f6e6c2202a4a48a" STYLE="fork"/>
            <node TEXT="无需重新计算哈希" ID="a6f0faeb87c2d32b71a866b7b4b3dbae" STYLE="fork"/>
          </node>
          <node TEXT="缺点" ID="2baf6873e0368e220ca2db7db99b1408" STYLE="fork">
            <node TEXT="分区数难以选择" ID="58b79e729ad08977e043f6fad07f1082" STYLE="fork"/>
            <node TEXT="数据量变化导致分区不均" ID="5205a348d58cb8abf38bd47f72456881" STYLE="fork"/>
          </node>
        </node>
        <node TEXT="动态分区" ID="3de4fc506df5c06c5f69027b9ae8539f" STYLE="fork">
          <node TEXT="分区大小达到阈值时自动分裂" ID="883917d691d4b716046c9babf37be20e" STYLE="fork"/>
          <node TEXT="优点" ID="887ff78814b0961cdfc24ede66563dba" STYLE="fork">
            <node TEXT="适应数据量变化" ID="40d10bd25dcaf59c5daf87de731908ec" STYLE="fork"/>
            <node TEXT="自动负载均衡" ID="cd14dc2727227227f76d2015b63d8a8a" STYLE="fork"/>
            <node TEXT="节点利用率高" ID="c38e9ef1225b2ad556dcaa752a83e373" STYLE="fork"/>
          </node>
          <node TEXT="缺点" ID="95876b972b326599c1cae61bb3cf91dc" STYLE="fork">
            <node TEXT="初始阶段分区少" ID="36b63203ae0ceba0ae575fb93525d880" STYLE="fork"/>
            <node TEXT="分裂/合并开销" ID="6bb77a0dffbbfd88b797fcd99d08dba3" STYLE="fork"/>
          </node>
        </node>
        <node TEXT="按节点比例分区" ID="ac7846fc8ad6cda30d1f0d1662f52113" STYLE="fork">
          <node TEXT="每个节点设置固定数量的分区" ID="08bac8108c3712cf70471f846ba53532" STYLE="fork"/>
          <node TEXT="优点" ID="67f9f682320665de00d5d5117ab6fae8" STYLE="fork">
            <node TEXT="每个节点负载相同" ID="8feda12157d275553a4db4add1be1797" STYLE="fork"/>
            <node TEXT="节点增减简单" ID="353ea93b79138cc4b5ab549302e45332" STYLE="fork"/>
          </node>
          <node TEXT="缺点" ID="303bb70e830da045879a58fcacbbd5a8" STYLE="fork">
            <node TEXT="分区大小不可控" ID="dc3938dfb1648b86fbaca89a8dfb85e9" STYLE="fork"/>
            <node TEXT="需要一致性哈希" ID="da236388ad7def1a2fa1bbb2ecd39d45" STYLE="fork"/>
          </node>
        </node>
      </node>
    </node>
    <node TEXT="请求路由" ID="413656355e752dfc2340f052718c2144" STYLE="bubble" POSITION="right">
      <node TEXT="路由问题" ID="abcc5a70854a57b27722b959e31ded58" STYLE="fork">
        <node TEXT="客户端如何知道数据在哪个节点" ID="5918db41f8822d0730c2447f5b7c4500" STYLE="fork"/>
        <node TEXT="策略" ID="2ecfae3d905ececb4ccaf7f5e8218123" STYLE="fork">
          <node TEXT="客户端路由" ID="c6f8198d4b77f4ae34963b5144e80684" STYLE="fork">
            <node TEXT="客户端知道分区分配，直接连接正确的节点" ID="fc4942a6d775446e9d3a14034f7fa194" STYLE="fork">
              <node TEXT="" ID="837c492b31ca530814b210a22887d85b" STYLE="fork"/>
            </node>
            <node TEXT="优缺点" ID="60e2ee00d98380c5dfc2a1d6e2568caa" STYLE="fork">
              <node TEXT="无中间层，低延迟" ID="c27072bd9eb8ed2c951ba047c9ba8540" STYLE="fork"/>
              <node TEXT="无代理开销，高性能" ID="7c9f7d3a4ef28e8542b2acccd4474f8a" STYLE="fork"/>
              <node TEXT="客户端复杂" ID="0f3bc8d246aa0053b45b41651be666c9" STYLE="fork"/>
              <node TEXT="映射更新困难" ID="25fd43d6ad42c309de0e7656c503c2d9" STYLE="fork"/>
            </node>
          </node>
          <node TEXT="路由层" ID="f58c8969b46757b181835498218bcae7" STYLE="fork">
            <node TEXT="专门的路由层负责转发请求" ID="044eee4e24c521a1922a43c3d660204a" STYLE="fork">
              <node TEXT="" ID="2ca3b9c2a50e024b6dfdf5920eca4af6" STYLE="fork"/>
            </node>
            <node TEXT="优缺点" ID="8f74af280cb8c0c4d4c8ef2865ab1e25" STYLE="fork">
              <node TEXT="客户端简单" ID="7039a0948873c5c0e3c1dff0185e0c74" STYLE="fork"/>
              <node TEXT="映射更新集中" ID="9ad6ca801e547a891afd38e01784004b" STYLE="fork"/>
              <node TEXT="有额外延迟" ID="8d62d8dd236d36cb2f8580dac1ce1e60" STYLE="fork"/>
              <node TEXT="路由层是单点故障" ID="77b05ca628dd22c5392a44636060979f" STYLE="fork"/>
            </node>
          </node>
          <node TEXT="节点协调" ID="24fdea384ac0b7f2bf85de34a2afafbc" STYLE="fork">
            <node TEXT="客户端可以连接任意节点，节点负责转发" ID="5c5df8232d7963cd78169beb646ac84a" STYLE="fork">
              <node TEXT="" ID="9f768f68fd0f623773db1bb94c194d60" STYLE="fork"/>
            </node>
            <node TEXT="优缺点" ID="d95a5f16f485347f45ce6a6f3606313b" STYLE="fork">
              <node TEXT="客户端极简" ID="ce1520a5aa3566a616bcc2b829a2b46e" STYLE="fork"/>
              <node TEXT="无单点故障" ID="2a12fa261ce5d62a43b78333f8ee148a" STYLE="fork"/>
              <node TEXT="节点对等，去中心化架构" ID="cf2005e82bc6a344b8fbf3145fb1be62" STYLE="fork"/>
              <node TEXT="节点通信开销" ID="71f68790452fce3f6853619a72454d5d" STYLE="fork"/>
              <node TEXT="一致性维护复杂" ID="9a3494b7b540c1842f00f9c4ff1dd096" STYLE="fork"/>
            </node>
          </node>
        </node>
      </node>
      <node TEXT="路由元数据管理" ID="f70f79d138467b466fd92513e38b847c" STYLE="fork">
        <node TEXT="问题：如何维护和更新分区映射？" ID="9728fe9fa05c3e2029d4cb19e442cbdf" STYLE="fork"/>
        <node TEXT="策略" ID="b55f8ece2fb7d96a46244f597460a8ff" STYLE="fork">
          <node TEXT="集中式配置服务" ID="3ac86459336bc5381dde19b19b93aa74" STYLE="fork">
            <node TEXT="zookeeper/etcd" ID="4099f0a40bb1138365be4444a4eaacbb" STYLE="fork"/>
            <node TEXT="优缺点" ID="9190f4df41939820a125a173fcad979a" STYLE="fork"/>
          </node>
          <node TEXT="Gosspip协议" ID="50b256d7e7b5f0a6eb0fe11dca90a9c9" STYLE="fork">
            <node TEXT="节点间定期交换元数据" ID="6c10b53f663e76d5587669ddec6725ef" STYLE="fork"/>
            <node TEXT="优缺点" ID="38ff6ac7c0e18f0f7abf9a15170f0c2f" STYLE="fork">
              <node TEXT="去中心化" ID="662deb04f1a9e05a06cd785c84d1391c" STYLE="fork"/>
              <node TEXT="自动收敛" ID="f4095339340b999dacb3cf1b353aeb02" STYLE="fork"/>
              <node TEXT="网络开销大" ID="8c69557f93601cb4ed7f88b64b77407a" STYLE="fork"/>
              <node TEXT="只有最终一致性" ID="6b8d977c0fdbe1fffb6da1601f808eed" STYLE="fork"/>
            </node>
          </node>
        </node>
      </node>
    </node>
    <node TEXT="分区的权衡和最佳实践" ID="876e4c7a816ed74794d9e7ed6ca07fdd" STYLE="bubble" POSITION="right">
      <node TEXT="" ID="af18584ecde30929acc74ed54ca11f25" STYLE="fork"/>
    </node>
  </node>
</map>