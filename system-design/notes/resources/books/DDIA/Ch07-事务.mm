
<map>
  <node ID="root" TEXT="Ch07-事务">
    <node TEXT="背景" ID="cc19fe90c50f66df71e677671b0a554a" STYLE="bubble" POSITION="right">
      <node TEXT="通过将程序的多个读、写操作作为一个操作逻辑单元，简化出错的情况" ID="7d3c231a72c071bfa37f280bd008e2b4" STYLE="fork"/>
      <node TEXT="核心思想：一组操作作为一个整体执行" ID="b0ff2ded1d54c15fa177913701b9e5b3" STYLE="fork"/>
    </node>
    <node TEXT="ACID详解" ID="0d67241c131852df50c5b9cd25407080" STYLE="bubble" POSITION="right">
      <node TEXT="原子性" ID="a5aa239c75ebb5cb1d7b5697fa1a755a" STYLE="fork">
        <node TEXT="事务中的所有操作作为一个不可分割的单元；要么全部成功，要么全部失败" ID="caba6b7cbc1bf7bea7ce7074ad77e19b" STYLE="fork"/>
        <node TEXT="实现原理：WAL（预写日志）" ID="162a2da7a9024f8715d60237e79df9f7" STYLE="fork">
          <node TEXT="在修改数据之前，先将修改操作写入日志" ID="efa2b2fa925391502c62f84de98ff662" STYLE="fork"/>
        </node>
      </node>
      <node TEXT="一致性" ID="68b9c472da177ba572ca7c4971881726" STYLE="fork">
        <node TEXT="数据库始终满足应用定义的不变式" ID="8ad13908b86f65a5aae4ec2e18e19107" STYLE="fork"/>
        <node TEXT="一致性的层次" ID="4201a77f2b08e6a3ba70365569efa00d" STYLE="fork">
          <node TEXT="数据库约束" ID="283bdec996603fd6f917f983268793fc" STYLE="fork">
            <node TEXT="主键、外键、唯一约束" ID="f3021d121c3ff26b6518baff18b07f0f" STYLE="fork"/>
            <node TEXT="CHECK约束" ID="9a0b8063af4993f94c31a828d0f1f764" STYLE="fork"/>
            <node TEXT="数据库自动检查" ID="58852f95d5636c56c4de263f6caee394" STYLE="fork"/>
          </node>
          <node TEXT="应用层不变式" ID="3b6828111b4144f03a65cde2c1eae258" STYLE="fork">
            <node TEXT="业务规则" ID="32573885e4dba43d0c21565a37737803" STYLE="fork"/>
            <node TEXT="需要应用逻辑保证" ID="f5c94d70542d9042082ab6da0dda5046" STYLE="fork"/>
          </node>
        </node>
      </node>
      <node TEXT="隔离性" ID="ebed542d87ddccd16109b479fa6d5185" STYLE="fork">
        <node TEXT="多个事务并发执行时，每个事务感觉自己是唯一在运行的事务" ID="f8f333a0bd20894d38b781a42bd08c36" STYLE="fork"/>
        <node TEXT="隔离级别" ID="abca052325afa611202173838ada312e" STYLE="fork">
          <node TEXT="Read Uncommited（读未提交）" ID="c8b735a274ada41738cc8353a6124140" STYLE="fork">
            <node TEXT="可以读到其他事务未提交的更改" ID="49487ac089b548b66bb64b72884b8a59" STYLE="fork"/>
            <node TEXT="最弱，几乎没有隔离性" ID="af60d8e15db476f3b74c25200a18f79f" STYLE="fork"/>
          </node>
          <node TEXT="Read Commited（读已提交）" ID="9e0073dd50cc9fa3d8caeb21487e48eb" STYLE="fork">
            <node TEXT="只能读到已提交的更改" ID="5d819f9a424a7189b52f856f33d5277d" STYLE="fork"/>
            <node TEXT="最常用的默认级别" ID="05b015954ef45f701b0dc55fe4f4cabe" STYLE="fork"/>
          </node>
          <node TEXT="Repeatable Read（可重复读）" ID="0f0c1bc1eee010a05b6fa0a98958a46b" STYLE="fork">
            <node TEXT="同一事务内多次读取同一数据，结果相同" ID="fd28d4866a35f43bd80aec96eb8785b4" STYLE="fork"/>
          </node>
          <node TEXT="Serializable（串行化）" ID="21ad4f266af4ff8a6335d93661189663" STYLE="fork">
            <node TEXT="最强隔离，等价于串行执行" ID="e59fc0c1da2c1f8f5a16758fcf248771" STYLE="fork"/>
            <node TEXT="性能最差" ID="c4d2afe033b29eb9e730c0ef9bb88d0d" STYLE="fork"/>
          </node>
        </node>
      </node>
      <node TEXT="持久性" ID="4623e784b54bbdf94b0c1a77da74c46f" STYLE="fork">
        <node TEXT="事务提交后，数据永久保存" ID="5bcb42b3216ca8bf6f4314da805e82c9" STYLE="fork"/>
        <node TEXT="WAL的持久性保证" ID="e4de59bb24f46f144e85fe6b733a61b7" STYLE="fork">
          <node TEXT="流程" ID="3e3603d747b1bdd432587ce3f7a157b1" STYLE="fork">
            <node TEXT="事务修改数据" ID="d96b85f3de4baa7395bd4350bee180a6" STYLE="fork"/>
            <node TEXT="写WAL到操作系统缓冲区" ID="848c64359e6730be63ce46371aaac407" STYLE="fork"/>
            <node TEXT="COMMIT时调用fsync" ID="ff1e769e939494c240a91a829d2ecdef" STYLE="fork"/>
            <node TEXT="返回客户端成功" ID="e5276bf34485df44bc11c1666b3923f0" STYLE="fork"/>
          </node>
        </node>
        <node TEXT="持久性的限制" ID="7604a4bf012b73e0ff4fe5a010c9cdbf" STYLE="fork">
          <node TEXT="硬件故障" ID="5d52871f291d5cfbe236dc8b2df00d2f" STYLE="fork"/>
          <node TEXT="数据损坏" ID="895c4c33c01e6b8d858a9a658b59e1d9" STYLE="fork"/>
          <node TEXT="人为错误" ID="843cd2f1e5daf3fa9ea0464c516c6eb7" STYLE="fork"/>
        </node>
      </node>
    </node>
    <node TEXT="并发控制问题" ID="a205b6fd28b22d75d8c8bfef9e0e074f" STYLE="bubble" POSITION="right">
      <node TEXT="脏读" ID="4eca8d38ca720a1bba95c90ce043d161" STYLE="fork">
        <node TEXT="读到其他事务未提交的数据" ID="a70ed648c50dad37556c2213a6bdfeb9" STYLE="fork"/>
        <node TEXT="设置Read Commited或者更高的隔离级别" ID="3d7533ca5e01f9dcb2c3e971fac9a065" STYLE="fork">
          <node TEXT="使用读锁" ID="88cdad13d1ea643233473b17cba2ca96" STYLE="fork"/>
          <node TEXT="使用MVCC，每个事务只能看到数据快照" ID="bbaf9706e8031bae2adc95b592fe68ae" STYLE="fork"/>
        </node>
      </node>
      <node TEXT="不可重复读" ID="f741150420a3d9e20495db03f5c7fd56" STYLE="fork">
        <node TEXT="同一事务内两次读取同一数据，结果不一致" ID="8fa29e8734ef6652201c46bd19a7421e" STYLE="fork"/>
        <node TEXT="设置Repeatable Read或Serializable隔离级别" ID="ae05893a754d5a1c5baf4e92e5edd22f" STYLE="fork">
          <node TEXT="使用MVCC快照隔离" ID="2ad10c58c9876c5c8039fbc7a6c6793d" STYLE="fork"/>
        </node>
      </node>
      <node TEXT="幻读" ID="9cb7a4db9ba4e4c176ed2b90b2f0e535" STYLE="fork">
        <node TEXT="同一查询返回不同的行集合（行数变换了）" ID="bce14b7ca3b76b3b22129b30007c27e9" STYLE="fork"/>
        <node TEXT="解决方案" ID="27b75ce9e7729818ce0ee40f8df8ae17" STYLE="fork">
          <node TEXT="使用谓词锁" ID="2045fa524c92fa6ed794114a37a954c7" STYLE="fork">
            <node TEXT="锁定查询条件范围" ID="de45609b91653dd4bb4c9f355302de42" STYLE="fork"/>
            <node TEXT="实现复杂，性能差" ID="78bffc7fba558490173677677d0df3c3" STYLE="fork"/>
          </node>
          <node TEXT="间隙锁" ID="edcdb488c6ad2efc8b89120cd44eb451" STYLE="fork">
            <node TEXT="锁定索引范围之间的“间隙”" ID="afd4d1b49c373b114342dd0693038ab1" STYLE="fork"/>
            <node TEXT="mysql innodb引擎采用" ID="86e35c929e1195bcd5a0e4a680e053f2" STYLE="fork"/>
          </node>
          <node TEXT="串行化快照隔离" ID="9efb149d874edceae6dfebe1a8c87d47" STYLE="fork">
            <node TEXT="检测到冲突时回滚一个事务" ID="10f2786b80baef95243574242b8aaa0b" STYLE="fork"/>
          </node>
        </node>
      </node>
      <node TEXT="丢失更新" ID="242677812289eb70bc5575bb8e6b86f6" STYLE="fork">
        <node TEXT="一个事务的更新覆盖了另一个事务的更新" ID="8b159920613cc79dd49d44dcf4c080dd" STYLE="fork"/>
        <node TEXT="解决方案" ID="b0344b52929957d3134c4079e77022fa" STYLE="fork">
          <node TEXT="使用数据库的原子操作，避免读-修改-写模式" ID="ef1bf14dfb284a97cec91c491fd38b02" STYLE="fork"/>
          <node TEXT="显式锁定" ID="2166b81769545be86ba8ea808fcdba94" STYLE="fork"/>
          <node TEXT="Compare-and-set" ID="b31410fe8c9a843c4b564c4585d93aad" STYLE="fork"/>
          <node TEXT="冲突解决；允许并发更新，后续合并冲突" ID="9d888b3e35115583bf9513cbb74a56c4" STYLE="fork"/>
        </node>
      </node>
      <node TEXT="写偏序" ID="88e2a5f454196bd044ae9d8d66daae6a" STYLE="fork">
        <node TEXT="两个事务读取相同数据，基于此做决策，并发写入导致不变式被破坏" ID="d3f2ea77d399af4b528716a46800e12b" STYLE="fork"/>
        <node TEXT="案例" ID="574b899297c11dc3738e266583bbeb64" STYLE="fork">
          <node TEXT="会议室预定" ID="ebfc14135c27f63e9e84f4a8858b3c67" STYLE="fork"/>
          <node TEXT="医生值班" ID="1b07f3a0ab33bef9fe4f9a0d9eb24576" STYLE="fork"/>
        </node>
        <node TEXT="解决方案" ID="08d2d42cc4f13b36b79c927060df83e3" STYLE="fork">
          <node TEXT="使用串行化隔离级别" ID="407ab9dcfba7d4011293926f789cae50" STYLE="fork"/>
          <node TEXT="显式锁定范围" ID="9b91487a75cd809a26f650e851851faa" STYLE="fork"/>
          <node TEXT="物化冲突" ID="8d952615c9b879abdddb5a7ac9dc8f3a" STYLE="fork">
            <node TEXT="创建一个表来显示跟踪约束" ID="5cb42282bc8f4e5768e8f5e3b8ca79d4" STYLE="fork"/>
          </node>
          <node TEXT="应用层锁" ID="e429c7c29dd0b24e01679a5c84d3f444" STYLE="fork"/>
        </node>
      </node>
    </node>
    <node TEXT="隔离级别详解" ID="bc014ab67b700bb3df2aaacac37b3081" STYLE="bubble" POSITION="right">
      <node TEXT="Read Uncommitted" ID="3b36bf18ea527f58ee9f1892216cc30d" STYLE="fork">
        <node TEXT="可以读到未提交的数据（脏读）" ID="5bdcc4e1fe942daa76dd1d05df739232" STYLE="fork"/>
        <node TEXT="最弱的隔离级别" ID="c48488b38d8fe460f5737d3b55424fde" STYLE="fork"/>
        <node TEXT="问题" ID="d8a251999f58cdfce613844d9535e1e3" STYLE="fork">
          <node TEXT="脏读" ID="2f48fcb5755401afa898bc6575eb4bdf" STYLE="fork"/>
          <node TEXT="不可重复读" ID="d68e7356e0d4a6826c51b1e7133ed0f6" STYLE="fork"/>
          <node TEXT="幻读" ID="8da0b7513100cf49c147102cf695a562" STYLE="fork"/>
          <node TEXT="丢失更新" ID="e01b15cecf44f4d59cceef424d84a99d" STYLE="fork"/>
        </node>
      </node>
      <node TEXT="Read Committed" ID="750cfbe6008e1685a844d8f275e1387e" STYLE="fork">
        <node TEXT="只能读到已提交的数据" ID="fb19d6a0fbce8d44b53689d568093434" STYLE="fork"/>
        <node TEXT="大多数数据库的默认级别" ID="b496d28bf6f36fb99dd976c3439d2ec4" STYLE="fork"/>
        <node TEXT="实现" ID="01d0b60b49912b235e4859f26d9e3dbf" STYLE="fork">
          <node TEXT="读取数据加共享锁；写入时加入排他锁" ID="ada5aa93cb51974cfa86bd30f6407e0c" STYLE="fork"/>
          <node TEXT="MVCC（更常用）" ID="f64a58b3416aef4a2aec3f7ca095cfb3" STYLE="fork"/>
        </node>
        <node TEXT="问题" ID="8ae0d5bc0b437812bdc13d154f94bfc5" STYLE="fork">
          <node TEXT="不可重复读" ID="b6e97f49e3ede727e8674114c117f889" STYLE="fork"/>
          <node TEXT="幻读" ID="a4b7ea11ecfc24951cc067e48f61342f" STYLE="fork"/>
          <node TEXT="丢失更新（需要额外的机制）" ID="d5eb471382677a17eef937f4ed26898c" STYLE="fork"/>
        </node>
      </node>
      <node TEXT="Repeatable Read" ID="b24166641447925316978c5420b97dd8" STYLE="fork">
        <node TEXT="同一事务内多次读取统一数据，结果相同" ID="a4d03a128f5337076c04036a53e0a880" STYLE="fork"/>
        <node TEXT="实现：MVCC快照隔离" ID="49f5e01198aa695d385f58cfff24492f" STYLE="fork"/>
        <node TEXT="问题" ID="ff1d3195184267e9bded45662349585e" STYLE="fork">
          <node TEXT="幻读（取决于实现）" ID="acd02f4c9299714bc592d5198429e1ad" STYLE="fork"/>
        </node>
      </node>
      <node TEXT="Serializable（串行化）" ID="9090948f7a0fd2e27aff948d3fc08b20" STYLE="fork">
        <node TEXT="特性" ID="391819815d4efe02bc8f7adec4efa649" STYLE="fork">
          <node TEXT="最强隔离级别" ID="51d247b31c5a28c60478f87ce739d111" STYLE="fork"/>
          <node TEXT="等价于串行执行" ID="51d687eeefae826a39ac333599a77d2f" STYLE="fork"/>
          <node TEXT="避免所有的并发问题" ID="2228b739a7707f55e44550e4bdb280cc" STYLE="fork"/>
        </node>
        <node TEXT="实现" ID="0fcc0aa2e95c5ddbc4ee854ec74a47d3" STYLE="fork">
          <node TEXT="单线程执行所有事务" ID="c197954c8934b5d9608ac1f2cc6ffc49" STYLE="fork"/>
          <node TEXT="两阶段锁" ID="3d11027ef582dc8b423d5a3dd44585a2" STYLE="fork"/>
          <node TEXT="串行化快照隔离" ID="88714f22a0f16a55acc6dfcb0ed0747e" STYLE="fork"/>
        </node>
        <node TEXT="问题" ID="76d9b298edb252b3aa5675fc87a6db67" STYLE="fork">
          <node TEXT="性能影响" ID="040f93b339f0592d7b0991b56ea51fbf" STYLE="fork"/>
        </node>
      </node>
      <node TEXT="快照级别对比" ID="5b99a3c7cf1513ad4f2eef36f1c681da" STYLE="fork">
        <node TEXT="" ID="981c193beef7cf98e2c39c77978dd908" STYLE="fork"/>
      </node>
    </node>
    <node TEXT="并发控制机制" ID="76afc9cc2ea99327fba8245a8828522b" STYLE="bubble" POSITION="right">
      <node TEXT="两阶段锁" ID="fa6f1ede143947fdb6d94783f70ba7ef" STYLE="fork">
        <node TEXT="分为加锁阶段和释放锁阶段" ID="cb878800c60f50cda962ecdbc24ec4c4" STYLE="fork"/>
        <node TEXT="规则" ID="958e1be550ecb6694b51e61697adcb74" STYLE="fork">
          <node TEXT="事务必须先获取锁才能读/写数据" ID="4844efb77f9875455bdf0b53647f059e" STYLE="fork"/>
          <node TEXT="事务不能再释放锁之后再获取新锁" ID="43bc734c0f09d6488360a79782da1a3e" STYLE="fork"/>
          <node TEXT="持有锁直到事务结束" ID="000206532e66048e13c918234533591f" STYLE="fork"/>
        </node>
        <node TEXT="锁类型" ID="d2c8ccde5fabc23395c1699f4778527b" STYLE="fork">
          <node TEXT="共享锁（s锁）" ID="99d4cc10d82283bc8ca58451cb330563" STYLE="fork">
            <node TEXT="读锁，多个事务可同时持有" ID="79101dade1445252b8a0fbcfbb9a226f" STYLE="fork"/>
          </node>
          <node TEXT="排他锁（x锁）" ID="0e33c796a430f7235d54f953aba27f29" STYLE="fork">
            <node TEXT="写锁，独占" ID="6b3b3645f3b2194721dab65e9e022e34" STYLE="fork"/>
          </node>
        </node>
      </node>
      <node TEXT="MVCC（多版本并发控制）" ID="b6ba06438577f08db3599c2dae5429eb" STYLE="fork">
        <node TEXT="维护数据的多个版本" ID="64c4464a38076862473e997526d08460" STYLE="fork">
          <node TEXT="写入时创建新版本，不覆盖旧版本" ID="3e40043ccac8cb848e8bc2a58bfd58d8" STYLE="fork"/>
          <node TEXT="读取时根据事务开始时间选择版本" ID="199832239d37523ce00c08407ec6b5b3" STYLE="fork"/>
          <node TEXT="读写不阻塞" ID="38a1a546c229f316fae4433523afb284" STYLE="fork"/>
        </node>
        <node TEXT="优点" ID="ec692d81a3e2a7b61e9de930d6ca6f6e" STYLE="fork">
          <node TEXT="读写不阻塞" ID="01c6a235668ea30e8871895483d7b2c2" STYLE="fork"/>
          <node TEXT="高并发" ID="a7c5a0de61f5a983074746439cbfbd64" STYLE="fork"/>
          <node TEXT="时间旅行查询（可以查询历史数据）" ID="01b55d321e3d6ca7c31a08d058558847" STYLE="fork"/>
        </node>
        <node TEXT="缺点" ID="5abf56c6c36d0e746e0ee1a349e10408" STYLE="fork">
          <node TEXT="空间开销大" ID="aaf61871b5ba95e322399faffa35cdad" STYLE="fork"/>
          <node TEXT="写放大" ID="81d714910697588aa00abcf2aaff6c4b" STYLE="fork">
            <node TEXT="更新=插入新版本+标记旧版本" ID="ecd76b674d81a74454af972d371de04c" STYLE="fork"/>
          </node>
          <node TEXT="长事务问题" ID="ed86ff676d7ade6e77d7cc0b6b102bc7" STYLE="fork">
            <node TEXT="长事务阻止旧版本清理" ID="8ef03b5d8b23ce799c00b8cfb8d524af" STYLE="fork"/>
            <node TEXT="磁盘空间膨胀" ID="c54cf001cbd8b7df509ccce9b3f251e7" STYLE="fork"/>
          </node>
        </node>
      </node>
    </node>
    <node TEXT="串行化快照隔离（SSI）" ID="0fb68736e5d496124381475d2f49965a" STYLE="bubble" POSITION="right">
      <node TEXT="在快照隔离的基础上检测冲突" ID="dcbf4d8e07fee8e43580a72701499152" STYLE="fork">
        <node TEXT="允许并发执行（乐观）" ID="56c286474eb13999a1ed68b98a6d5fb5" STYLE="fork"/>
        <node TEXT="检测不可串行的模式" ID="954283c22a0b72a33b751e2fd3a2b6a0" STYLE="fork"/>
        <node TEXT="发现冲突后回滚" ID="ba8f8d114031b68e47fd06f4f28505dc" STYLE="fork"/>
      </node>
      <node TEXT="优点" ID="74e201473556bdee5d0cd2aae6dab416" STYLE="fork">
        <node TEXT="高性能" ID="b2af78a514262407a43d060bf06ca545" STYLE="fork"/>
        <node TEXT="保证串行化" ID="f3449eb296f6c103e9331ffecccc8582" STYLE="fork"/>
        <node TEXT="无死锁" ID="7dbd44e4658e8ebc27a5516a07f8a33a" STYLE="fork"/>
      </node>
      <node TEXT="缺点" ID="bfd1514901b3ff031cfac340e5eb3b8d" STYLE="fork">
        <node TEXT="可能回滚" ID="5d2b849de4418d2232367e03422de159" STYLE="fork"/>
        <node TEXT="高冲突时性能下降" ID="f503db08c0c7a128cab10b74581e29be" STYLE="fork"/>
        <node TEXT="实现复杂" ID="481b69b131b3d8116b787b8d7ecb70f2" STYLE="fork"/>
      </node>
    </node>
    <node TEXT="主流数据库隔离级别与实现方案对比" ID="dc96d78ddba8e1ff1e0b3e9314b89e05" STYLE="bubble" POSITION="right">
      <node TEXT="隔离级别支持与默认配置对比" ID="64ff9383c798254ea253f170588c072a" STYLE="fork"/>
      <node TEXT="RC实现机制" ID="0eb0249c8c8daaebeefddc2ff792199b" STYLE="fork"/>
      <node TEXT="RR实现机制" ID="eed42f9dc299fe3c770d89370904c21c" STYLE="fork"/>
      <node TEXT="串行化实现机制" ID="21702681a9300a29d8285f65bddea0a2" STYLE="fork"/>
    </node>
  </node>
</map>