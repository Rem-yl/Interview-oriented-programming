
<map>
  <node ID="root" TEXT="Ch05-数据复制">
    <node TEXT="背景" ID="be5be1fcf700a93aaec598426801e7f5" STYLE="bubble" POSITION="right">
      <node TEXT="复制主要指通过互联网络在多台机器上保存相同数据的副本" ID="9abdd646d3eabec9769eeb647d76b9c1" STYLE="fork"/>
      <node TEXT="期望达到的目的" ID="51ad4a92786608c2a964a3de14e6162e" STYLE="fork">
        <node TEXT="使数据在地理位置上更接近用户，从而降低访问延迟" ID="86b7e90912e71ab4104ec6aff9067a2c" STYLE="fork"/>
        <node TEXT="当部分组件出现故障，系统依然可以继续工作，提高可用性" ID="d6282035ef48e80b18219e3ee3dbf5ab" STYLE="fork"/>
        <node TEXT="扩展至多台机器以同时提供数据访问服务，从而提高吞吐量" ID="b840f689a838b26f4216364f2096583c" STYLE="fork"/>
      </node>
      <node TEXT="本章暂不讨论数据分区的情况" ID="23db268c26d1c52a32aea21c7fd12932" STYLE="fork"/>
      <node TEXT="核心问题" ID="92560f4b991955f3a983f5ce65ecd36e" STYLE="fork">
        <node TEXT="如何处理那些持续更改的数据" ID="afb5c5dc6a78cd08a293a2ca8aa7f56c" STYLE="fork"/>
        <node TEXT="如何确保所有副本之间的数据是一致的？" ID="e32a0487fa9691497bc9e2dce5f59fda" STYLE="fork"/>
      </node>
      <node TEXT="主流方法" ID="b2f3bed657b117a64bd1247d05b8950b" STYLE="fork">
        <node TEXT="主从复制" ID="377cc28c82678feee97058006659c176" STYLE="fork"/>
        <node TEXT="多主节点复制" ID="eddb3616b64bd0dd36eb7bd3dded8bf1" STYLE="fork"/>
        <node TEXT="无主节点复制" ID="81d35df2992bc51e3b5cd9bd6f5bda96" STYLE="fork"/>
      </node>
    </node>
    <node TEXT="主从复制" ID="5c0054c08be81ddc5bbfc339301d2864" STYLE="bubble" POSITION="right">
      <node TEXT="核心思想" ID="4807ba74112da85592651e95a441ae19" STYLE="fork">
        <node TEXT="指定一个副本作为主节点（leader）" ID="dd1ef4be728efa4dc8da76ef1af946ec" STYLE="fork"/>
        <node TEXT="其他副本作为从节点（follwer）" ID="7c9b4d77790bdb54ec5a79dd88cafe27" STYLE="fork"/>
        <node TEXT="所有写入请求都发送到主节点" ID="e66e737a14c1ade797bc854c3a8e7cda" STYLE="fork"/>
        <node TEXT="主节点将数据变更以复制日志（Replication Log）的形式发送给从节点" ID="5d7037abb8bc41d8d299d0d3862e45c1" STYLE="fork"/>
        <node TEXT="从节点应用变更，保持和主节点的数据一致" ID="229d6af905049aa1e2aa6fe8e9941083" STYLE="fork"/>
      </node>
      <node TEXT="同步复制和异步复制" ID="88ca7ff9a1969ae6ab493caf669c57e8" STYLE="fork">
        <node TEXT="同步复制" ID="4a30028b30be20e23c0f39b33beb79d1" STYLE="fork">
          <node TEXT="主节点等到至少一个从节点确认写入成功之后，才向客户端报告成功" ID="b9718b122fe02b067948c6afecf9f480" STYLE="fork"/>
          <node TEXT="优点" ID="76e43397d5d97da44daee3f3b893925c" STYLE="fork">
            <node TEXT="从节点保证拥有与主节点一致的最新数据副本" ID="f764f82475bb981890efbdd872db78b6" STYLE="fork"/>
            <node TEXT="如果主节点失效，可以确保数据不会丢失" ID="584e67d65f6ec10cf6f760ee32f0d541" STYLE="fork"/>
          </node>
          <node TEXT="缺点" ID="1eb4f2687e70ce809b5d9e7d3f51480d" STYLE="fork">
            <node TEXT="如果同步的从节点无法响应，写入就无法完成" ID="8262f7705bfe691811ef5c0960ca63b8" STYLE="fork"/>
            <node TEXT="主节点必须阻塞所有写入，等到不可用的从节点恢复" ID="2a90b1025d470751b2037049e944fa78" STYLE="fork"/>
            <node TEXT="写入延迟增加" ID="fdafdb2f60da6def792f203c03028c07" STYLE="fork"/>
          </node>
        </node>
        <node TEXT="异步复制" ID="1dd8a6048f3d2013a7f278a21d7ab9d4" STYLE="fork">
          <node TEXT="主节点发送写入请求后，不等待从节点确认就向客户端报告成功；从节点在后台异步应用更改" ID="40b698d05dc964478460ffcabfda247e" STYLE="fork"/>
          <node TEXT="优点" ID="c7b650d1e55491822ccbe926ac49f243" STYLE="fork">
            <node TEXT="主节点可以继续处理写入，即使所有从节点都落后" ID="e803aba2a93d844058d8d422fbf217ed" STYLE="fork"/>
            <node TEXT="写入延迟低" ID="c842ea009f4f2a3686368805463beacd" STYLE="fork"/>
          </node>
          <node TEXT="缺点" ID="3c9683cb43dea14ec24a9dcde0fb774d" STYLE="fork">
            <node TEXT="如果主节点失效，写入尚未复制到从节点，则写入会丢失" ID="72870e140e95c00a6a35e85b223b25f7" STYLE="fork"/>
            <node TEXT="数据持久性较弱" ID="1b16825072c38fa35dbd9bb0cecb9dbb" STYLE="fork"/>
          </node>
        </node>
        <node TEXT="半同步复制" ID="31dd25a0e1053e286557ffdc24401fee" STYLE="fork">
          <node TEXT="一个从节点是同步的，其他从节点是异步的" ID="4e16b55f0e8ea52fb6ca6386945292d6" STYLE="fork"/>
          <node TEXT="如果同步从节点不可用，则将要给异步从节点提升为同步" ID="d7ad65c76b018de92290048a26a64659" STYLE="fork"/>
          <node TEXT="保证至少有两个节点拥有最新数据副本" ID="fb5069dac3a853e40c0e50a0b6ceec22" STYLE="fork"/>
        </node>
      </node>
      <node TEXT="配置新的从节点" ID="89f4907f06896f267149a1d610189e93" STYLE="fork">
        <node TEXT="获取主节点快照" ID="5871886131681c6bcba2403eccd19bf5" STYLE="fork"/>
        <node TEXT="将快照复制到新的从节点" ID="9f27f88875c346757039d3bcf28cafa7" STYLE="fork"/>
        <node TEXT="连接到主节点并请求快照点之后的所有数据变更" ID="16ce4464040288f7685beb5290f23818" STYLE="fork"/>
        <node TEXT="从节点处理积压的数据变更" ID="efeaf1729efd594f45c30e9aae5e65d5" STYLE="fork"/>
      </node>
      <node TEXT="处理节点失效" ID="da52407624f3e694f2bf15380b95998f" STYLE="fork">
        <node TEXT="从节点失效：追赶式恢复" ID="692c754d5885faac5412afa7b657772a" STYLE="fork">
          <node TEXT="从节点崩溃后，从本地磁盘的复制日志中知道故障前处理的最后一个事务" ID="cbaa59cfb081a57c534c589b3f29769b" STYLE="fork"/>
          <node TEXT="从节点重新连接到主节点" ID="692d71b747438913913e71f549067441" STYLE="fork"/>
          <node TEXT="请求故障期间发生的所有数据变更" ID="8d41ad724594576053aa588a325aa0d3" STYLE="fork"/>
          <node TEXT="应用这些变更，追赶上主节点" ID="955387ac4cc267d38266747ad6d6920a" STYLE="fork"/>
        </node>
        <node TEXT="主节点失效：故障转移" ID="b3dac6eca92038e39e80e1a9c1140276" STYLE="fork">
          <node TEXT="解决流程" ID="2eaf6e9166e4f41252e9c6fc64214b81" STYLE="fork">
            <node TEXT="确定主节点失效" ID="a835fcc90f54eac224c33f3e0d81142e" STYLE="fork">
              <node TEXT="没有绝对可靠的方法检测节点失效" ID="9bf1f2c87bbc0a58476bf9803b64cb49" STYLE="fork"/>
              <node TEXT="通常使用超时机制" ID="e457c4b12e0fc0018068bb732e64dd18" STYLE="fork"/>
            </node>
            <node TEXT="选举新的主节点" ID="7fc0f01a187a0d63ee8ef66098197710" STYLE="fork">
              <node TEXT="通过选举过程（多数节点同意）" ID="2dfae4d6bebd12489eae24ea33dd2842" STYLE="fork"/>
              <node TEXT="或者由预先指定的控制节点任命" ID="6f44d7201aa31cbc5a89fe23db4e4eeb" STYLE="fork"/>
              <node TEXT="最佳候选者通常是拥有最新数据变更的从节点" ID="2ee5908c266e898752d0235b01413ec3" STYLE="fork"/>
            </node>
            <node TEXT="重新配置系统以使用新主节点" ID="8e14548b46506e56f484a5074b59ff4f" STYLE="fork"/>
          </node>
          <node TEXT="主要挑战" ID="f8b8072a241e854336f5cbb3b6c707a0" STYLE="fork">
            <node TEXT="异步复制的数据丢失问题" ID="6619f844eeccdd8aa756a4a4b0e98370" STYLE="fork"/>
            <node TEXT="脑裂问题" ID="599bc53e6e6448304e8bdbcc71fdee09" STYLE="fork">
              <node TEXT="旧主节点恢复后可能仍认为自己是主节点" ID="b59c2f5c8dedbde68cae5c483cf7ad47" STYLE="fork"/>
            </node>
            <node TEXT="超时时间设置" ID="3ddcd52ac4e469604b830d58dd85c451" STYLE="fork"/>
            <node TEXT="与其他系统的交互" ID="a56ce18f1cdefa61074d30e7c483b376" STYLE="fork"/>
          </node>
        </node>
      </node>
      <node TEXT="复制日志的实现" ID="70c2c3a2898bd8dc10985cb5cc245476" STYLE="fork">
        <node TEXT="基于语句的复制" ID="8e80a468db4ec26924e72131da917a04" STYLE="fork">
          <node TEXT="原理" ID="c21eae2ac2e4721933d8f9604aa7031f" STYLE="fork">
            <node TEXT="主节点记录每个写入请求（语句）并发送给从节点" ID="eddb77db250a3867e94f7d42e2dd1701" STYLE="fork"/>
            <node TEXT="从节点解析并执行SQL语句" ID="40140d36ef56cada13dbb8b49860f30d" STYLE="fork"/>
          </node>
          <node TEXT="问题" ID="898e664990517ee88c7cdfc5f851f8be" STYLE="fork">
            <node TEXT="非确定性函数" ID="7909ba4f53d69be16f15d7cff872d3f6" STYLE="fork">
              <node TEXT="NOW() RAND()函数在每个副本上会产生不同的值" ID="7cbd5b151718d6a3ea4b4e2b00308b57" STYLE="fork"/>
            </node>
            <node TEXT="自增列或依赖现有数据的语句" ID="2ec9ca40e4f8dd3f9d22bd3461917277" STYLE="fork"/>
            <node TEXT="有副作用的语句" ID="9268b735990f4469dc3649d73121815c" STYLE="fork">
              <node TEXT="触发器、存储过程、用户定义函数等可能产生不同的副作用" ID="c6de0373cb72515ab0cbb107d86eae7f" STYLE="fork"/>
            </node>
          </node>
        </node>
        <node TEXT="基于预写日志（WAL）传输" ID="7a5b2ac4809e0778e9e96a1709061f1d" STYLE="fork">
          <node TEXT="原理" ID="b2dfd838de9f3fea632f9d410af4df37" STYLE="fork">
            <node TEXT="存储引擎通常使用预写日志" ID="87bfce5101198d5d57c821c7aa8bb384" STYLE="fork"/>
            <node TEXT="日志是包含所有数据库写入的仅追加字节序列" ID="460a03444290aa75af892b4bfaebd418" STYLE="fork"/>
            <node TEXT="主节点除了将日志写入磁盘，还将其发送给从节点" ID="fceb8f0c60bcd3952fb9894ee6f80033" STYLE="fork"/>
            <node TEXT="从节点处理日志，构建与主节点完全相同的数据结构副本" ID="a806aad73046b50e6d27151e81d1ccbc" STYLE="fork"/>
          </node>
          <node TEXT="缺点" ID="e5e85440bfa9bcf9815186cc130cf2f0" STYLE="fork">
            <node TEXT="WAL描述的数据是非常底层的：涉及磁盘块的具体字节" ID="40522f3c1d5f617830d8ec1bdcb7c508" STYLE="fork"/>
            <node TEXT="复制与存储引擎紧密耦合" ID="9de300a4a2c4a6a7fbe07fc32a7c6f5b" STYLE="fork"/>
            <node TEXT="如果数据库版本不同，WAL格式可能不兼容" ID="347402d9e71790e78405bb67cb246f18" STYLE="fork"/>
            <node TEXT="升级系统时需要停机" ID="7ad37c859eea75daa3a98d3d8c1244ea" STYLE="fork"/>
          </node>
        </node>
        <node TEXT="基于行的逻辑日志复制" ID="6e58a78687bd8281b8119d565ae49ee8" STYLE="fork">
          <node TEXT="原理" ID="ba6bfa445b88d92c1af48b21c172f2ba" STYLE="fork">
            <node TEXT="使用不同的日志格式进行复制和存储" ID="701937495e8be37ddc9c9f302eae0a80" STYLE="fork"/>
            <node TEXT="复制日志与存储引擎解耦，称为逻辑日志" ID="2f619d1a42c283dd04628b890c14e4e1" STYLE="fork"/>
          </node>
        </node>
        <node TEXT="基于触发器的复制" ID="e93a37b9b854060515e81c7917dca574" STYLE="fork">
          <node TEXT="原理" ID="68563bececa9a9ce71c590f3c9ea9ff3" STYLE="fork">
            <node TEXT="使用应用层代码" ID="380ede0c3deabd05354da065fab12cc0" STYLE="fork"/>
            <node TEXT="数据库触发器：当数据变更时自动执行自定义代码" ID="7b7ac4a2708458d2d899e97f43ce39cb" STYLE="fork"/>
            <node TEXT="触发器可以将数据变更记录到单独的表中" ID="9f7868cb2fed2d83d6d61b7640e63cca" STYLE="fork"/>
            <node TEXT="外部进程读取该表，应用必要的业务逻辑，并将变更复制到其他系统" ID="5cb62b5fc4a1892a1a08df59e0376f5a" STYLE="fork"/>
          </node>
          <node TEXT="缺点" ID="677bb14dbaeb166bad332d1c5ed2bd32" STYLE="fork">
            <node TEXT="比其他复制方法有更大的开销" ID="abeed437db27ffda14a75194eb01df32" STYLE="fork"/>
            <node TEXT="更容易出错" ID="5d27c22cc459e3d9d378a2098c5cbf65" STYLE="fork"/>
            <node TEXT="有更多限制" ID="5e72bd2ea4fd23822cba0dbaad557bbd" STYLE="fork"/>
          </node>
        </node>
      </node>
      <node TEXT="复制延迟问题" ID="93fd6ffd95f604f393b4b10706abe1e9" STYLE="fork">
        <node TEXT="最终一致性" ID="b78d702c615ae0e7d421bfcb48a3869c" STYLE="fork">
          <node TEXT="概念" ID="fc8a4cbca90f724a4cacc33ca832adf3" STYLE="fork">
            <node TEXT="异步复制系统中，从节点落后于主节点" ID="4cdebebd6dd87aa65d488b321698edd0" STYLE="fork"/>
            <node TEXT="如果停止写入，从节点最终会赶上并与主节点一致" ID="7d55f3cae2f204f5f4da65d0e39ef036" STYLE="fork"/>
          </node>
          <node TEXT="复制延迟" ID="631dd3c7c68126d21ffbbb75a7caf9b0" STYLE="fork">
            <node TEXT="延迟可能只是几分之一秒" ID="e214577b2dd929267e68fd9076d46477" STYLE="fork"/>
            <node TEXT="在高负载或网络问题时，可能达到几秒甚至几分钟" ID="aacf5dff596c3fc9a2cb3d4107d9cadc" STYLE="fork"/>
            <node TEXT="延迟过大会导致应用层面的实际问题" ID="031803da8e3bb42988545a55eab62608" STYLE="fork"/>
          </node>
        </node>
        <node TEXT="读自己的写" ID="414c29f99ed3f8e0d26dd4f5325de685" STYLE="fork">
          <node TEXT="问题场景" ID="523e0ab1a81151b0463879042bbaa960" STYLE="fork">
            <node TEXT="用户提交数据之后立即查看" ID="10f77740b6259d94ca2b52ca02fe7c43" STYLE="fork"/>
            <node TEXT="新数据发送到主节点，但是从异步从节点读取" ID="52a77fe5ae15435529434cf58fb1927a" STYLE="fork"/>
            <node TEXT="如果复制延迟大，新数据可能还未到达从节点" ID="bbbe69df6dd5176a3c1dfa7a59ed729b" STYLE="fork"/>
            <node TEXT="用户会认为刚提交的数据丢失了" ID="57cd8e255d213f17b52b8377f4d6c3cc" STYLE="fork"/>
          </node>
          <node TEXT="解决方案" ID="36722a07e99e23239f282c24f0e407ef" STYLE="fork">
            <node TEXT="读写一致性" ID="e4dd0b10672f19b232ec042c203980ac" STYLE="fork">
              <node TEXT="读取用户可能修改过的内容时，从主节点读取" ID="4c12070b117aea0e3d150cb2024d515e" STYLE="fork"/>
              <node TEXT="跟踪最后更新时间" ID="e2d438414be3fd4776cec247e4bdad12" STYLE="fork"/>
              <node TEXT="客户端记住最近写入的时间戳" ID="1d345c85aa43cd69994a7bfba57df3de" STYLE="fork"/>
              <node TEXT="跨设备的读写一致性" ID="67fd09df0101a244f5bffc48e950635e" STYLE="fork">
                <node TEXT="元数据需要集中化" ID="8607332d13b44c0fec4a59ca9651d9b0" STYLE="fork"/>
                <node TEXT="不同设备的请求可能路由到不同的数据中心" ID="6592874d516596f90196fe5193398e05" STYLE="fork"/>
              </node>
            </node>
          </node>
        </node>
        <node TEXT="单调读" ID="63867fac4de58bd10319f9de605f5f18" STYLE="fork">
          <node TEXT="问题场景" ID="453c0c37d9bfb5e37b19452f496bbbf1" STYLE="fork">
            <node TEXT="用户多次读取，可能连接到不同的从节点" ID="441ea9085a787fbf660c7378fd4118e8" STYLE="fork"/>
            <node TEXT="用户可能会看到时光倒流：第二次读取的数据比第一次读取的旧" ID="0cd689399bbdb73a3b6b92f4f9ba21aa" STYLE="fork"/>
          </node>
          <node TEXT="解决方案" ID="3f654b4ffd1ab0005c6fa73ade87c170" STYLE="fork">
            <node TEXT="单调读一致性" ID="024a16a4de5b86647e9d22f35ef24a4e" STYLE="fork">
              <node TEXT="确保用户总是从同一个从节点读取" ID="b80b8d865bb009510407a1f4457d5d09" STYLE="fork"/>
              <node TEXT="如果该从节点失效，则重新路由到另一个" ID="2f771e0d168d3fb33da48a20b8a5db01" STYLE="fork"/>
            </node>
          </node>
        </node>
        <node TEXT="前缀一致读" ID="48e9029aa0e96f1513210c0d2a6235db" STYLE="fork">
          <node TEXT="问题场景" ID="f70267e2ef3d8107cec1339670440631" STYLE="fork">
            <node TEXT="一系列按特定顺序发生的写入" ID="ccf1f856b7cc9ecd78fa0b37e85d09da" STYLE="fork"/>
            <node TEXT="读取这些写入的人，也应该按照相同的顺序看到它们" ID="9e10eaabde616b8121c3feafd86e008a" STYLE="fork"/>
            <node TEXT="在分区数据库中尤为重要" ID="0d8a7b7d66f13951e7fbea5872925576" STYLE="fork">
              <node TEXT="不同的分区以不同的速度运行" ID="781d2d914c2558de23152d32bd11bde9" STYLE="fork"/>
              <node TEXT="没有全局的写入顺序" ID="344e4bc8bcb4364a357af6711a5e4357" STYLE="fork"/>
            </node>
          </node>
          <node TEXT="解决方案" ID="4d390eea579be4cde9cd9a4ce35a89fc" STYLE="fork">
            <node TEXT="确保有因果关系的写入发送到同一个分区" ID="ee82e368502da634fc8ed761d78586e1" STYLE="fork"/>
            <node TEXT="在某些应用中，高效的跟踪因果依赖关系很困难" ID="31eb5a8b87f192680939ac5f86529e77" STYLE="fork"/>
          </node>
        </node>
      </node>
      <node TEXT="主从复制的应用场景" ID="5d85c47d5ae26ed226f651217a26a3d2" STYLE="fork">
        <node TEXT="读多写少的应用" ID="e4b3f7e857df2cd164ed52caad4a321a" STYLE="fork"/>
        <node TEXT="需要高可用性的系统" ID="746b793b997c9c1ca3433a0ef107379c" STYLE="fork"/>
        <node TEXT="地理式分布系统" ID="2e723d46bee41744fa266dbe4c9ace49" STYLE="fork"/>
        <node TEXT="MySQL, PostgreSQL, Kafka, MongoDB ...." ID="c45a5526108301efca5c7375bf9e76ad" STYLE="fork"/>
      </node>
    </node>
    <node TEXT="多主节点复制" ID="78e99aa408322c1399f273bcb14f1e5f" STYLE="bubble" POSITION="right">
      <node TEXT="单主节点复制的局限性" ID="e04cb4fab93e0fee8eb4cf917be8c804" STYLE="fork">
        <node TEXT="单点故障：Leader挂了就无法写入" ID="e37736ec0325fbb538b2f26d61968bd1" STYLE="fork"/>
        <node TEXT="跨数据中心延迟：所有写入都要到一个数据中心" ID="763709688b96ec047c01c64a90d31855" STYLE="fork"/>
        <node TEXT="离线操作困难：客户端必须连接到leader" ID="c8cd1b9bf132c037f2862ec65d9bfa87" STYLE="fork"/>
        <node TEXT="性能瓶颈：所有写入压力集中到一个节点" ID="e3827892ba91bf50d3779969aa68ecfa" STYLE="fork"/>
      </node>
      <node TEXT="使用场景" ID="041d32ff1be6b9b3576b93db53ce27a6" STYLE="fork">
        <node TEXT="多数据中心运维" ID="267db6e24433397b4feb8775fa1bb0f7" STYLE="fork"/>
        <node TEXT="离线操作的客户端" ID="d5389bf73709f7a001e3013674943a0f" STYLE="fork">
          <node TEXT="在线文档的离线编辑" ID="d90af9d4110d2323b1433fd3fe13361d" STYLE="fork">
            <node TEXT="用户在本地离线修改文档（本地leader）" ID="1d2e05267db4257435690612bfe5c8dd" STYLE="fork"/>
            <node TEXT="网络恢复和同步到云端（远程leader）" ID="46626010388db9e71d27fcd027ee7db7" STYLE="fork"/>
            <node TEXT="其他设备从云端拉取最新数据" ID="a21755dd76eafb8c48e8566b80cf0245" STYLE="fork"/>
          </node>
        </node>
        <node TEXT="协同编辑" ID="fc5dc905b44336261a666cc8b9ffe387" STYLE="fork">
          <node TEXT="每个用户都是一个leader" ID="14044f87dd7142585ed7f11846b4a060" STYLE="fork"/>
          <node TEXT="本地及时响应用户输入" ID="d1aa7d5c9f08f3b53b98f976967c9a29" STYLE="fork"/>
          <node TEXT="使用CRDT或OT算法解决冲突" ID="a91ff22570b7e18b63314011b7af69ac" STYLE="fork"/>
        </node>
      </node>
      <node TEXT="核心挑战" ID="40b31d224aa004225a3e6647b60c2b2a" STYLE="fork">
        <node TEXT="写冲突" ID="30bf88bcf1d01ff7a68f78d78e71a346" STYLE="fork">
          <node TEXT="根本原因：不同用户同时修改同一条数据" ID="0228a07a021c37111517155eea1ba5cb" STYLE="fork"/>
        </node>
        <node TEXT="冲突检测" ID="27bbe31cd9270f87dc3d57167ccf738a" STYLE="fork">
          <node TEXT="同步冲突检测" ID="bfdb0e6b1b6a1e784573791e33dddc4d" STYLE="fork">
            <node TEXT="写入时检测冲突" ID="70a77ce68bc3b8d229f593cbe4b08ac3" STYLE="fork"/>
            <node TEXT="性能退化到单主节点" ID="cf8e595554bf83e0207768b5e9af75a8" STYLE="fork"/>
          </node>
          <node TEXT="异步冲突检测" ID="c13ae942562aadb4d0813eae79c7dfc4" STYLE="fork">
            <node TEXT="先写入，后检测冲突" ID="f6522bfa92556975f1c152c18e488cb9" STYLE="fork"/>
            <node TEXT="多主节点标准做法" ID="bf25dceb7c28a7463376679d5ed34102" STYLE="fork"/>
          </node>
          <node TEXT="解决策略" ID="ad9966f1abee5ffda136f550b5f9c526" STYLE="fork">
            <node TEXT="避免冲突" ID="4053abdc0e56f3c500e980a91b747838" STYLE="fork">
              <node TEXT="数据分区" ID="9d529b1d980747834749477dedd263fe" STYLE="fork">
                <node TEXT="规则：用户始终写入同一个数据中心" ID="7c5d6b47669145a508a21875314b201d" STYLE="fork"/>
              </node>
              <node TEXT="数据归属" ID="299c7bd4ece97e88ba2a9e0c9ab177ea" STYLE="fork">
                <node TEXT="规则：文档只在所有者的主数据中心可写" ID="bfcc67a7a8095f9c46c7536475d637a8" STYLE="fork"/>
                <node TEXT="协作编辑困难" ID="63f8de1d0fbb413280b0a19a8e9fd87a" STYLE="fork"/>
              </node>
            </node>
            <node TEXT="收敛到一致的状态" ID="b5e5746ef3715f27afdfbbdeabf6987b" STYLE="fork">
              <node TEXT="最后写入获胜" ID="e29de734b32dc833a4fdf25854c2224c" STYLE="fork">
                <node TEXT="原理：给每个写入附加时间戳，时间戳最大的获胜" ID="e851c9d2f8b3d5b894e18d07ce9a49b3" STYLE="fork"/>
                <node TEXT="存在的问题" ID="07433af7aa8cbbd7dbba977dded26a36" STYLE="fork">
                  <node TEXT="时间戳较早的修改可能完全丢失" ID="89807eb1076932efbac03695e418aa6b" STYLE="fork"/>
                  <node TEXT="始终不可靠，不同服务器的时钟可能不同步" ID="f0d72c3c5649f31d8086a13d5f1b40e4" STYLE="fork"/>
                  <node TEXT="因果关系丢失" ID="220fe18dfa808f8c570a712088941775" STYLE="fork"/>
                </node>
              </node>
              <node TEXT="基于版本向量的合并" ID="368935677c19f4aa6a6c7d1f58c3bce7" STYLE="fork">
                <node TEXT="原理：跟踪每个副本的向量，识别因果关系" ID="7683d539e6385b90ecc1f3c9256c8267" STYLE="fork"/>
              </node>
              <node TEXT="自定义冲突解决逻辑" ID="bd3be5d78a1fac0739403abbe8128aae" STYLE="fork">
                <node TEXT="数据库返回冲突数据，由应用程序决定" ID="e5c4863ca0384852301a722c7a95ea33" STYLE="fork"/>
              </node>
              <node TEXT="CRDT" ID="a1f7bf8c84b2ee00cf5a1ec167f65eff" STYLE="fork">
                <node TEXT="设计无冲突的数据结构" ID="ec125f674ce3bad6683065270816eb50" STYLE="fork"/>
              </node>
            </node>
          </node>
        </node>
      </node>
      <node TEXT="多主节点复制拓扑" ID="478c24781e1542f8987a28afc0d896bf" STYLE="fork">
        <node TEXT="环形拓扑" ID="bdcc7a1a92b7372648e28b6f487743c1" STYLE="fork">
          <node TEXT="每个节点只连接到下一个节点" ID="3c23abf7481ba121951d23989ccfccaa" STYLE="fork"/>
          <node TEXT="连接数少" ID="e8788cbf5f854fa5c855c326cdd79b77" STYLE="fork"/>
          <node TEXT="单点故障，任一节点挂了，环断裂" ID="f1ec0063058d5f54cdf5eaea6912d3ba" STYLE="fork"/>
          <node TEXT="延迟高，需要经过多个节点" ID="a11403ed8cf51b728b72da55eff29d40" STYLE="fork"/>
        </node>
        <node TEXT="星形拓扑" ID="a08f3f105b81151e56865bbf40d916d9" STYLE="fork">
          <node TEXT="一个根节点连接其他所有节点，根节点转发所有写入" ID="965d58fe032faf4158ef92e12118bb6f" STYLE="fork"/>
          <node TEXT="延迟低" ID="4a093baede275441c1b5203f00004535" STYLE="fork"/>
          <node TEXT="根节点是单点故障" ID="8c8e59d3f8c2df07f0e506cbd8ea2193" STYLE="fork"/>
          <node TEXT="根节点性能瓶颈" ID="6e2f025e9def4c775a86a0f931fc8106" STYLE="fork"/>
        </node>
        <node TEXT="全连接拓扑" ID="d0f7752b3d816b1f6bcdf4371a349ea4" STYLE="fork">
          <node TEXT="每个节点连接到其他所有节点" ID="e93487922d6eb2b5a940d784c4f77fc3" STYLE="fork"/>
          <node TEXT="写入直接发送到所有节点" ID="1b684dfe644e9f21d505845cac807eb2" STYLE="fork"/>
          <node TEXT="无单点故障，延迟最低" ID="89b467d0f5ce1396f5cb92e0d951bd46" STYLE="fork"/>
          <node TEXT="连接数多且冲突检测复杂" ID="af5ba34f407a246dcf2875f6f007563b" STYLE="fork"/>
        </node>
      </node>
      <node TEXT="优点" ID="819140fc6db4a103f026b68ee69d7355" STYLE="fork">
        <node TEXT="性能好" ID="f6e5b6349655611f2f4d54fa8ff02df4" STYLE="fork"/>
        <node TEXT="可用性强" ID="3e7a78e6f5a1b49fe19dae6fb7db0f84" STYLE="fork"/>
        <node TEXT="容错性高" ID="8b2915996aa20ce7526d6caf0506d513" STYLE="fork"/>
        <node TEXT="支持离线操作" ID="5960158ef77570c4bae6ffb9d6b56147" STYLE="fork"/>
      </node>
      <node TEXT="缺点" ID="c899e5376097b1a8bcf506b3ce9f126a" STYLE="fork">
        <node TEXT="冲突检测复杂，调试困难" ID="837a8af7d3d7e9a862559f08347df6e4" STYLE="fork"/>
        <node TEXT="最终一致性，不是强一致性" ID="396e1039c374dec809148164fd6afe36" STYLE="fork"/>
        <node TEXT="运维成本高" ID="925c8533aa0a718df72b8e57f68bc731" STYLE="fork"/>
        <node TEXT="应用层处理复杂" ID="4a03e5bfbf3fdd3b454b1896bc4586fa" STYLE="fork"/>
      </node>
    </node>
    <node TEXT="无主节点复制" ID="d72b4b33307366b63a866fe08b337749" STYLE="bubble" POSITION="right">
      <node TEXT="核心思想：没有leader，所有节点地位平等" ID="6b57f03139c977bc84f8da99b9a325ea" STYLE="fork"/>
      <node TEXT="工作原理" ID="24ace2c71b50c628101331eab867cc72" STYLE="fork">
        <node TEXT="写入流程" ID="c797c5b4b80d17723f10f88d6a0e895b" STYLE="fork">
          <node TEXT="并行写入多个副本" ID="af79c20fd9da307f3a536aaec288fceb" STYLE="fork"/>
          <node TEXT="客户端发送写请求到协调者节点" ID="8b03d465089681f57f32a9df97c237d1" STYLE="fork"/>
          <node TEXT="协调者并行写入到N个副本节点" ID="29dae28422d7a79f642b0cec97430e9e" STYLE="fork"/>
          <node TEXT="等待W个节点确认写入成功" ID="2fdc25ba2a7bdad36dd133395110d664" STYLE="fork">
            <node TEXT="W = 写仲裁数" ID="ebe562ba51abc0eb516e587bd729516d" STYLE="fork"/>
            <node TEXT="仲裁条件：W + R &gt; N" ID="e3b1516cce871080b23d7c90d39591e5" STYLE="fork"/>
            <node TEXT="N：总副本数；W：写仲裁数；R：读仲裁数" ID="758b4e54c2fdeb68e0399e20929c22dd" STYLE="fork"/>
          </node>
          <node TEXT="返回成功给客户端" ID="e35df13c93eddd8536d22bbc9a9d97a9" STYLE="fork"/>
        </node>
        <node TEXT="读取流程" ID="ae3299be520e7b8746164f7836cbabac" STYLE="fork">
          <node TEXT="并行读取多个副本" ID="74ed278b8ad31f4e02129b9288d0f3aa" STYLE="fork"/>
          <node TEXT="客户端发送读请求到协调者" ID="ea2ce81610f341ea7badc6e312bc3825" STYLE="fork"/>
          <node TEXT="协调者并行从R个节点读取" ID="688f81e9dd0e283af488f442866d275f" STYLE="fork"/>
          <node TEXT="比较版本号，选择最新的值" ID="b15c0ed736533ed9512438a045e3227d" STYLE="fork"/>
          <node TEXT="后台修复过期数据" ID="6035eba32fe006db0f79f7d1a890fc38" STYLE="fork"/>
        </node>
        <node TEXT="仲裁机制" ID="ac77799fdc82f185b3e31de412e6ed72" STYLE="fork">
          <node TEXT="W+R&gt;N" ID="dbb4237c7c3c634e46f34d1f870c6381" STYLE="fork">
            <node TEXT="至少有一个节点同时参与了读和写" ID="51a57769930c10d73c1ecb5f1785449d" STYLE="fork"/>
            <node TEXT="保证能读到最新写入的数据" ID="d24e308947af5b6bd127b032a48dcdaa" STYLE="fork"/>
          </node>
          <node TEXT="常见配置" ID="3e3f76f35623e287c474ea3b1c524b22" STYLE="fork">
            <node TEXT="" ID="32e095deaa9d38129e2c98d6e83a6698" STYLE="fork"/>
          </node>
          <node TEXT="仲裁一致性的局限" ID="62c0561a0e875044eb803a53650ed113" STYLE="fork">
            <node TEXT="仲裁不保证强一致性" ID="cfc24824836739869fab6d8662b911a0" STYLE="fork"/>
            <node TEXT="问题1：并发写入" ID="23a0bb19391bd0f880f61d17affb59d8" STYLE="fork"/>
            <node TEXT="问题2：网络分区" ID="fecac9b29a48201c5eb9bacff3a7c4e0" STYLE="fork"/>
            <node TEXT="问题3：时间戳不可靠" ID="274de37d4d95800d35a56b2a23d67615" STYLE="fork"/>
          </node>
        </node>
        <node TEXT="数据修复机制" ID="2d52e65d653adfa8254938aa23b00f70" STYLE="fork">
          <node TEXT="读修复：读取时发现数据不一致，顺便修复" ID="192705f53a1adeac0549636f34b85574" STYLE="fork"/>
          <node TEXT="反熵过程：后台进程定期同步数据" ID="04bc22d58f4abaad51f54a2fbfa4af6f" STYLE="fork">
            <node TEXT="每个节点定期与其他节点比较数据" ID="ce370354d4e685350588a211b8c63cc4" STYLE="fork"/>
            <node TEXT="对比方式：Merkle Tree" ID="59a176fe212e4837658d9e4fe95c4432" STYLE="fork"/>
            <node TEXT="比较根哈希：不同则递归比较子树，找到差异块" ID="5bdae71598561f92571bb80f2478d589" STYLE="fork"/>
            <node TEXT="只同步差异数据" ID="a0a5a8ab7bdc3c6536f4532aaaa69580" STYLE="fork"/>
          </node>
          <node TEXT="提示移交：解决节点临时故障" ID="c6da7923b062b78b295e3a2070ce4c1d" STYLE="fork">
            <node TEXT="将数据临时存储到其他节点" ID="8494d91031e4af7ce81284fe5b2aa325" STYLE="fork"/>
            <node TEXT="待故障节点恢复后移交数据" ID="8f37cfc0bf40b1071fbc5068c61cd4c3" STYLE="fork"/>
          </node>
        </node>
        <node TEXT="版本控制和冲突检测" ID="56921bf75db89ee52d78d8f064528a4d" STYLE="fork">
          <node TEXT="版本号" ID="a702cdacd6e723bcf34d972caed5f96d" STYLE="fork">
            <node TEXT="单一版本号无法检测并发写入" ID="adb82cd56440fe01bab018c9ba15a8c3" STYLE="fork"/>
          </node>
          <node TEXT="版本向量" ID="4125078fe8dadc952cf9968090446c66" STYLE="fork">
            <node TEXT="每个副本独立记录版本号" ID="0dc54c138990b0509e2d0362569316f9" STYLE="fork"/>
            <node TEXT="可以检测并发写入" ID="a56a949152ea5cdd5627092888e944bd" STYLE="fork"/>
            <node TEXT="由应用层解决冲突" ID="2ea6120af0eb821efe2c41ff0eb50572" STYLE="fork"/>
          </node>
          <node TEXT="Dotted Version Vectors" ID="7bdf3afe8a9c4cc4b6eacb3bf22a9f6a" STYLE="fork">
            <node TEXT="解决版本向量的空间开销问题" ID="dd5d38474e3b74628d06854a3feed953" STYLE="fork"/>
          </node>
        </node>
        <node TEXT="宽松仲裁" ID="748ade5006123bd37d024caac0dd5b20" STYLE="fork">
          <node TEXT="解决传统仲裁节点故障导致的写入失败问题" ID="020d5d6615e30c2a1329aeecffafe432" STYLE="fork"/>
          <node TEXT="通过临时使用其他节点凑够W，等故障节点恢复之后再转移数据" ID="09dacbd0dd9b32b9c339ae911395bf74" STYLE="fork"/>
        </node>
      </node>
      <node TEXT="优点" ID="67fec44c44156fc878629486a423081d" STYLE="fork">
        <node TEXT="高可用" ID="7019b71007018ff69b34feed6d49f8e4" STYLE="fork"/>
        <node TEXT="可扩展" ID="b1168916437bf51a7f8c782c32d98c05" STYLE="fork"/>
        <node TEXT="去中心化" ID="aca483645262d084909d31187b844bfc" STYLE="fork"/>
        <node TEXT="低延迟" ID="f5889137b00bc3aeef00d89cc3014299" STYLE="fork"/>
        <node TEXT="跨数据中心友好" ID="5e41d44f35db5f11bfca130d97ceb52d" STYLE="fork"/>
      </node>
      <node TEXT="缺点" ID="129ff7612b2962aa6ebded11dce5fb93" STYLE="fork">
        <node TEXT="应用层需要处理不一致" ID="7b7799a2e3f51c340293742f9dae20b0" STYLE="fork"/>
        <node TEXT="冲突解决复杂" ID="b79f2e4a9a35de6bb103433268da8c57" STYLE="fork"/>
        <node TEXT="读修复开销" ID="15e2d5f495f63098b2ab1a1dd54fe5f6" STYLE="fork"/>
        <node TEXT="元数据开销" ID="d4031a3916f45c2f9ffd855b0ff1762b" STYLE="fork"/>
        <node TEXT="运维复杂" ID="bfe617c1be030914594d30dda5291a97" STYLE="fork"/>
      </node>
    </node>
  </node>
</map>