
<map>
  <node ID="root" TEXT="Ch08-分布式系统的挑战">
    <node TEXT="核心观点" ID="5a64b0491d1dfb2fdf763d02b7e8a59c" STYLE="bubble" POSITION="right">
      <node TEXT="分布式系统中的故障是常态，而非异常" ID="0d164cb06c1c3c275fd3a6aa5fa9434c" STYLE="fork"/>
      <node TEXT="网络、时钟、进程都是不可靠的" ID="38906a559ed0e414c0dab114b1db97e2" STYLE="fork"/>
      <node TEXT="必须明确系统的假设和保证" ID="4daceb34deac1ec29915abb74ea06554" STYLE="fork"/>
    </node>
    <node TEXT="故障与部分失效" ID="09d01917a8fb04743d57f39d032a36a3" STYLE="bubble" POSITION="right">
      <node TEXT="单机系统与分布式系统" ID="111a7fce7a181aed760fe50ffe1d551f" STYLE="fork">
        <node TEXT="单机系统的特点" ID="b5dd73c386b411917f0232baca25cbcd" STYLE="fork">
          <node TEXT="确定性；程序要么全部执行，要么崩溃" ID="41cc5fc0b035bdb8e96782dcb86d5aea" STYLE="fork"/>
          <node TEXT="不会出现部分执行的情况" ID="9725d2064075e6205fc5d1fc9e6a6b8f" STYLE="fork"/>
        </node>
        <node TEXT="分布式系统的特点" ID="f94dde13293d7cee2c2166887dd5529d" STYLE="fork">
          <node TEXT="非确定性；部分组件可能失效，其他组件继续工作" ID="e87e6da70338546e87092317960d6370" STYLE="fork"/>
          <node TEXT="部分失效是常态" ID="fa74a84b5fca4e6b88fc08ce61f277b8" STYLE="fork"/>
        </node>
      </node>
      <node TEXT="构建可靠系统的方法" ID="8feadf6cd301b3ecb4b7b31d0159817d" STYLE="fork">
        <node TEXT="核心原则：在不可靠的组件上构建可靠的系统" ID="b9f1e3c1677cdf20c4383b2f1a0d475c" STYLE="fork"/>
        <node TEXT="机制" ID="5328644209e13c04f6c9afa52ecc52a4" STYLE="fork">
          <node TEXT="检测故障" ID="ae6c688aee67fe4659501573013b6324" STYLE="fork"/>
          <node TEXT="容忍故障" ID="741e9a96d1bf560d2b69cbd62c86f3da" STYLE="fork"/>
          <node TEXT="隔离故障" ID="28821f3e6760eabacf7a47549caf9707" STYLE="fork"/>
          <node TEXT="从故障中恢复" ID="8e027b2f79a9dc354f308527417d3766" STYLE="fork"/>
        </node>
      </node>
    </node>
    <node TEXT="不可靠的网络" ID="aa0f009fc780dbec6d16c9bd940c08d7" STYLE="bubble" POSITION="right">
      <node TEXT="网络故障的现实" ID="479c3b783c7c8271b9d462a9563e8938" STYLE="fork">
        <node TEXT="节点无法判断远程节点是因为故障了还是网络传输问题才没有得到响应" ID="411c9442cabb16febcbec42afe0bcc7b" STYLE="fork"/>
      </node>
      <node TEXT="故障检测" ID="4b6322192fb467487bf4927889ee1650" STYLE="fork">
        <node TEXT="无法区分这些情况" ID="ee515c3fd4c37359dfe900a0fcd76297" STYLE="fork">
          <node TEXT="节点崩溃" ID="63b3456aab59a278b5f55457d29aac3b" STYLE="fork"/>
          <node TEXT="节点宕机" ID="3d2573993000e64c89d36e5428fdb6e2" STYLE="fork"/>
          <node TEXT="网络故障" ID="8165736013fb816b21929b8427cf6ac0" STYLE="fork"/>
          <node TEXT="节点负载过高（缓慢响应）" ID="f9ed1f512e9027b206b10b96003ac19e" STYLE="fork"/>
        </node>
        <node TEXT="方案" ID="a37851aaa161e718d090f9027d8d35b3" STYLE="fork">
          <node TEXT="超时" ID="681797a4a11dfbfba21524e5856bb042" STYLE="fork"/>
          <node TEXT="心跳检测" ID="0577622cbdc272e398412f3fab8cb22a" STYLE="fork">
            <node TEXT="依然依赖超时" ID="df5fa758aef189323e0710950be1f382" STYLE="fork"/>
            <node TEXT="无法区分节点是崩溃还是网络故障" ID="85662ed6c60d25fef8e82075b1834f53" STYLE="fork"/>
          </node>
          <node TEXT="反向探测" ID="4bfd079de5439ef8b6960fd6426b6994" STYLE="fork">
            <node TEXT="给出怀疑值" ID="1e0dc63e820b6770680d601e21dee6b6" STYLE="fork"/>
            <node TEXT="根据历史响应时间动态调整" ID="b9b2ba9304885bb9e29d59bff491c83c" STYLE="fork"/>
          </node>
        </node>
      </node>
      <node TEXT="超时与重试" ID="99b73f781389d605b4327c7263640a80" STYLE="fork">
        <node TEXT="重试的复杂性" ID="c1d4c00ba57a63268821479aaa45f88d" STYLE="fork">
          <node TEXT="幂等性" ID="1393cb89bee8e0e07b0e168c888151fe" STYLE="fork">
            <node TEXT="需要保证多次操作的结果一致" ID="50e851d581b05858c60f8b90e28626e2" STYLE="fork"/>
            <node TEXT="解决方案" ID="377eba8eb856320ecf813cb78f14b9c7" STYLE="fork">
              <node TEXT="使用唯一请求ID" ID="d74a35327c362cf3cf4d7c0d4c77fdd3" STYLE="fork"/>
              <node TEXT="记录已处理的请求" ID="0206b57d21ede44a0c137c99a13607cb" STYLE="fork"/>
              <node TEXT="重复请求-&gt;返回缓存结果" ID="ccf4be0b7b55fec6565c5c8bb86e539f" STYLE="fork"/>
            </node>
          </node>
          <node TEXT="雪崩效应" ID="c271d1db7bc997b80c89f083adac291d" STYLE="fork">
            <node TEXT="服务器负载高时响应慢，更多的重试导致服务器负载更高" ID="b626d7856fed01938b4531f19ee126da" STYLE="fork"/>
            <node TEXT="解决方案" ID="1b4012e1207580c9182f5440ede7aa90" STYLE="fork">
              <node TEXT="退避重试" ID="9aa23f997b6c942aa176761e3e1b31c4" STYLE="fork"/>
              <node TEXT="限流" ID="7e12291bd248a2b61a859fb2dab7f2a7" STYLE="fork"/>
              <node TEXT="断路器" ID="cf6bf0e70eaea3349676114d06b2df7e" STYLE="fork"/>
            </node>
          </node>
        </node>
      </node>
      <node TEXT="网络拥塞与排队" ID="6cd5433efd20851f0d421d4a006ded97" STYLE="fork">
        <node TEXT="网络延迟的不稳定" ID="79e9262ff09d7c5efece39583560d835" STYLE="fork">
          <node TEXT="光速限制" ID="66f6ec09b25fcc7673ed7a352f17112e" STYLE="fork"/>
          <node TEXT="排队延迟" ID="4b93163ff8f066248e4cba54e83ae500" STYLE="fork">
            <node TEXT="发送方操作系统排队等待发送" ID="12ca83024421b6c8532776cd0084b5ec" STYLE="fork"/>
            <node TEXT="网络交换机排队发送网络包" ID="24bd890cd56e2ae0436d06544174ef4b" STYLE="fork"/>
            <node TEXT="接收方操作系统等待接收" ID="c10ed5925f548a20426e570fccead8dc" STYLE="fork"/>
            <node TEXT="接收方应用等待应用处理队列" ID="60cedd326d49707048d4642900acfc01" STYLE="fork"/>
          </node>
          <node TEXT="网络拥塞" ID="4edbb4fdbfe2c713cfa9379edbcde482" STYLE="fork"/>
        </node>
      </node>
    </node>
    <node TEXT="不可靠的时钟" ID="5a4e5c15e75aaea64425de5060e31d0e" STYLE="bubble" POSITION="right">
      <node TEXT="时钟的类型" ID="fe1122e5f08b21dfeaaaf0ff4ca89aa0" STYLE="fork">
        <node TEXT="墙上时钟" ID="aee908926258cf4184d28eceafd5fbc9" STYLE="fork">
          <node TEXT="用于表示日期和时间" ID="26eaa40754777ab3c382372b98d20b8e" STYLE="fork"/>
          <node TEXT="不适合测量时间间隔" ID="4bb47210d68878c29f9eca4aa237464f" STYLE="fork"/>
        </node>
        <node TEXT="单调时钟" ID="cf5c7b48dd658839cc69ba2ee81d24d9" STYLE="fork">
          <node TEXT="测量时间间隔" ID="b77849587f589bf6117860c07db7226b" STYLE="fork"/>
          <node TEXT="无法转换成日期时间；不同机器无法比较" ID="3996f5ddbec166b6d8f8dbbaa8fb4576" STYLE="fork"/>
        </node>
      </node>
      <node TEXT="时钟同步与漂移" ID="2bf3d030aab55262196798dd7ec5a7c7" STYLE="fork">
        <node TEXT="为什么需要时钟同步：计算机时钟不准确" ID="af946efa2ed059f8decae54c559623a6" STYLE="fork">
          <node TEXT="计算机时钟不准确" ID="439d22af54fd0599ee5142f58551c7c5" STYLE="fork"/>
        </node>
        <node TEXT="NTP同步原理" ID="c85562a8bde113d290fc7f76ce40d5a7" STYLE="fork">
          <node TEXT="可以简单理解成系统通过访问NTP服务器，计算时间偏移计算来矫正自己的时钟" ID="853605271e7ddee9b42f23f48105eaf6" STYLE="fork"/>
        </node>
      </node>
      <node TEXT="逻辑时钟" ID="f76d8b5fefc920d34fc00b02e5a5d40a" STYLE="fork">
        <node TEXT="基于事件的因果关系，而非物理时间" ID="9b32a3d08396f385b8e6948425b87ef2" STYLE="fork"/>
        <node TEXT="Lamport时间戳" ID="e44996ad7c06e57eb9f749a779b47cf4" STYLE="fork">
          <node TEXT="核心思想：如果事件A发生在事件B之前，那么A的时间戳大于B" ID="19c22695d06b5484609adeeb1b4aef60" STYLE="fork"/>
          <node TEXT="算法规则" ID="c3f203574b6ec572ebfa61b61fa8c5ec" STYLE="fork">
            <node TEXT="每个节点维护一个计数器" ID="e8a737230c61f125099f3a137f979fc3" STYLE="fork"/>
            <node TEXT="节点处理本地事件时:" ID="b852c5f4e4564f92153385351a82074a" STYLE="fork"/>
            <node TEXT="发送消息时" ID="467d31a9eaf7f085de61a4e7f575db2d" STYLE="fork"/>
            <node TEXT="接收消息时" ID="cf596e44e3810c0125658619fd07729f" STYLE="fork"/>
          </node>
          <node TEXT="性质" ID="b337ff6f57d4cc8ab027730ad4d0945c" STYLE="fork">
            <node TEXT="因果顺序保证" ID="52f65f27bbf1902dcefa1b5c08028558" STYLE="fork">
              <node TEXT="如果a-&gt;b，则L(a) &lt; L(b)" ID="e1113f00160e2c1db07787709ee20358" STYLE="fork"/>
            </node>
            <node TEXT="逆命题不成立" ID="fdc36cf3743c12655c302aab4687b922" STYLE="fork">
              <node TEXT="如果L(a) &lt; L(b)，推导不出a-&gt;b" ID="adb24a829d3929343f65720d1bcb933d" STYLE="fork"/>
            </node>
          </node>
        </node>
        <node TEXT="版本向量" ID="50b216cc838a342ec9d27eb196f9eb26" STYLE="fork">
          <node TEXT="lamport时间戳的局限性：无法检测并发事件" ID="fe324ff754e2f09834b83eb37a720ae1" STYLE="fork"/>
          <node TEXT="核心思想：每个节点跟踪所有节点的事件版本" ID="2503c85ce1189d5beb37c925b6f6bbde" STYLE="fork"/>
          <node TEXT="数据结构" ID="04f522ccbdef97b82b0689de6db23880" STYLE="fork">
            <node TEXT="vector = {node1:2, node2:1, node3:0}" ID="b1baddb94d4a8d1c6bc0ac4764a3e95b" STYLE="fork"/>
          </node>
        </node>
      </node>
    </node>
    <node TEXT="进程暂停" ID="aeeadb013f0fb527877096f1400636be" STYLE="bubble" POSITION="right">
      <node TEXT="背景" ID="2f69722ea454578b1b91ab61519c92b6" STYLE="fork">
        <node TEXT="问题：即使网络正常、时钟同步，节点进程也可能意外暂停" ID="e65f7ebe55143bc8e7f98b3d09859b83" STYLE="fork"/>
        <node TEXT="场景" ID="6ba642621d423ac681b0467b289ee2f4" STYLE="fork">
          <node TEXT="GC暂停" ID="21873205194e41c74a201409d636af8b" STYLE="fork"/>
          <node TEXT="虚拟机暂停/迁移" ID="e1a9e29f83323f3c4bd383afd6d500e9" STYLE="fork"/>
          <node TEXT="操作系统调度" ID="e9fd0b4b15eaff9abf1be732c08ba5c6" STYLE="fork"/>
          <node TEXT="磁盘I/O阻塞" ID="80943168ac6124b304bc81693ed6d147" STYLE="fork"/>
        </node>
        <node TEXT="进程暂停的影响" ID="307c0f71113c33910557e387512d36a1" STYLE="fork">
          <node TEXT="脑裂问题" ID="2d018b7e7bd22039d09509b15151e683" STYLE="fork">
            <node TEXT="GC暂停时暂停所有的应用线程，只进行垃圾回收" ID="59e8bfe405a2932114471a52e4e31688" STYLE="fork"/>
            <node TEXT="从外部看进程失联了；内部看只是暂停了一下" ID="319c8857fe3344a01252839d8f41d229" STYLE="fork"/>
          </node>
        </node>
      </node>
      <node TEXT="策略" ID="84f1ef305fe6312dfc5a930239c545da" STYLE="fork">
        <node TEXT="Fencing Token" ID="e4f0888ddfa1b7da752a4c5d893ac971" STYLE="fork">
          <node TEXT="不依赖进程的自我判断；由外部系统保证安全性" ID="37b16934f54b50965f84132a03f7b29b" STYLE="fork"/>
          <node TEXT="本质就是通过写入token，后续进程通过比较token大小来判断自己是否落后" ID="bb624dd10884ae0b68f91aeaab13deb0" STYLE="fork"/>
        </node>
        <node TEXT="避免长时间暂停" ID="327450df8a327af374764e27a630f9e6" STYLE="fork">
          <node TEXT="GC优化" ID="77dd812a50f47e2629c0c4ffeaae0cac" STYLE="fork"/>
          <node TEXT="检测暂停" ID="8aef311d731cf0c6327416ca6ac5bf40" STYLE="fork"/>
          <node TEXT="租约续期分离" ID="96ef4b0b5651be31bae7f88c85f4aadd" STYLE="fork"/>
          <node TEXT="实时操作系统" ID="da390c1ec018a24719024c25fb4db235" STYLE="fork"/>
        </node>
      </node>
    </node>
    <node TEXT="知识与真相" ID="b1b2837bd140e368ff7a37c246a89e50" STYLE="bubble" POSITION="right">
      <node TEXT="背景" ID="96b1908428266023a9fe07449fce7295" STYLE="fork">
        <node TEXT="网络不可靠-&gt;无法知道消息是否到达" ID="199f0083487b8c532e61415d20796735" STYLE="fork"/>
        <node TEXT="时钟不可靠-&gt;无法准确判断事件顺序" ID="7ad4fbc265609109802c9b5f358b91af" STYLE="fork"/>
        <node TEXT="进程不可靠-&gt;节点不知道自己发生了什么" ID="1b0b23614f048db5d056c528e9e4f852" STYLE="fork"/>
        <node TEXT="在分布式系统中，如何知道真相" ID="8f268c00565e35d6bc61be60d9cb32c4" STYLE="fork"/>
      </node>
      <node TEXT="分布式系统没有全局真相" ID="a37ca6c4ef8a93abfbb6be4414df64d1" STYLE="fork">
        <node TEXT="只有局部视角，没有全局真相" ID="348e9da16d6e2eb4ee31409adbd39336" STYLE="fork"/>
      </node>
      <node TEXT="拜占庭故障" ID="8c98d96187dece217545176a3f338d99" STYLE="fork">
        <node TEXT="问题" ID="5d48e7160b5f1830115159dab3b07131" STYLE="fork">
          <node TEXT="节点发送错误数据" ID="15852cca81a762e9190a2350fb3e7789" STYLE="fork"/>
          <node TEXT="给不同节点发送矛盾信息" ID="3c3c16e5db09a48b405a694bbf511ef1" STYLE="fork"/>
          <node TEXT="恶意破坏" ID="b1183eb343305d2e60a1a319990c6b03" STYLE="fork"/>
        </node>
        <node TEXT="拜占庭容错" ID="7e915c7d14d106beb130e024cc615a78" STYLE="fork">
          <node TEXT="需要n &gt;= 3f + 1个节点才能容忍f个拜占庭故障" ID="96e0d8eede27433ccde360e1b3017be8" STYLE="fork"/>
          <node TEXT="n是总节点数，f是拜占庭节点数" ID="3ded3ea26857854939b8efc434f7064d" STYLE="fork"/>
          <node TEXT="容错算法" ID="e9c444e52fcf6a26fa7f373ec6376e34" STYLE="fork">
            <node TEXT="PBFT" ID="86254b34791944afbe894477e7ac3354" STYLE="fork"/>
            <node TEXT="Bitcoin" ID="3a79d0f5aadb9f6cafebe59cdbd33e3b" STYLE="fork"/>
            <node TEXT="Ethereum 2.0" ID="63402c7b76aca04b3ea68c49ae26121c" STYLE="fork"/>
          </node>
        </node>
        <node TEXT="使用场景" ID="7778734c76112ea567d34a44758e6b2b" STYLE="fork">
          <node TEXT="比特币、航天行业这种需要极度保证安全的系统" ID="05c9834e84a0840e15a25136dd988e31" STYLE="fork"/>
        </node>
      </node>
    </node>
    <node TEXT="构建可靠的分布式系统" ID="56e1f23f25f04a227fc832576a068cfc" STYLE="bubble" POSITION="right">
      <node TEXT="设计原则" ID="7bf6d0037a8d6bf333708c2c594cae80" STYLE="fork">
        <node TEXT="假设一切都会失败" ID="d1765f5cb8b833d01ad04d3e52ca24fd" STYLE="fork"/>
        <node TEXT="端到端原则" ID="16433c30ef40c0e414689dd3d88de2d8" STYLE="fork">
          <node TEXT="可靠性保证应该在端到端层面实现" ID="411548c7252983766b97bb3554df50ca" STYLE="fork"/>
          <node TEXT="中间层的可靠性保证往往不够" ID="0702590d9e43cd9c7ee32b84d4c87346" STYLE="fork"/>
        </node>
        <node TEXT="幂等性" ID="64e1ad4ec36b5aa5df128633910bf1f9" STYLE="fork">
          <node TEXT="多次执行与执行一次结果相同" ID="d944813d98e1ec6ba68f179559b43aca" STYLE="fork"/>
        </node>
        <node TEXT="最终一致性+冲突解决" ID="3064e0cfdc2bfc29121d6dd37f8f8c00" STYLE="fork"/>
        <node TEXT="监控与可观测性" ID="87701407810e2455b3e92202e391bf2c" STYLE="fork"/>
      </node>
    </node>
  </node>
</map>