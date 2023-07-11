<script>
  import {AppStart, AppStop, ReadClientsMsg,GetClientsList, ReadRemoteMsg, CheckAppStop} from '../wailsjs/go/main/App.js'

  let name

  let cliDisPlaySelectAll = 'ALL'

  let startOrStop = false
  let targetIp = '127.0.0.1';
  let targetPort = '1231';
  let listenIp = '0.0.0.0';
  let listenPort = '1232';
  let selectedClient = cliDisPlaySelectAll;
  let clientOptions = [];
  let isPrinting = true; // add this line to track the printing state

  let infoPrint1 = '';  // declare a variable to store print info
  let infoPrintElement1;  // add this line to declare the variable
  let infoPrint2 = '';  // declare a variable to store print info
  let infoPrintElement2;  // add this line to declare the variable

  let isModalOpen = false;
  let modalContent = ""; // 新的状态变量

  let showServerMsgTask
  let showClientMsgTask
  let checkAppStopTask
  let updateClientSelectTask

  let isHex = false

  let clientCount = 0; // 连接的客户端数量

  function updateClientOptions(options) {
    if (clientOptions !== options){
      clientOptions = options;
      clientCount = options.length
    }
  }

  function getCurrentDatetime() {
    let currentDatetime = new Date();

    let year = currentDatetime.getFullYear();
    let month = String(currentDatetime.getMonth() + 1).padStart(2, '0'); // months are 0-indexed in JS
    let date = String(currentDatetime.getDate()).padStart(2, '0');

    let hours = String(currentDatetime.getHours()).padStart(2, '0');
    let minutes = String(currentDatetime.getMinutes()).padStart(2, '0');
    let seconds = String(currentDatetime.getSeconds()).padStart(2, '0');
    //精确到毫秒
    let milliseconds = String(currentDatetime.getMilliseconds()).padStart(3, '0');

    return `${year}-${month}-${date} ${hours}:${minutes}:${seconds}:${milliseconds}`;
  }

  function showModal(content) {
    modalContent = content; // 设置 modalContent 的值
    isModalOpen = true; // 打开模态弹窗
  }

  // a function to be called when you get some results
  function handleResult1(title, context) {
    if (!isPrinting) return; // add this line to check the printing state

    let lineFeed = ""

    if (infoPrint1 !== '') {
      lineFeed = '\n'
    }
    let shouldScroll = infoPrintElement1.scrollTop + infoPrintElement1.clientHeight === infoPrintElement1.scrollHeight;

    if (isHex){
      //context转化为16进制显示
      let hexStr = ''
      for (let i = 0; i < context.length; i++) {
        hexStr += context.charCodeAt(i).toString(16) + ' '
      }
      context = hexStr
    }

    infoPrint1 += lineFeed + '[' + getCurrentDatetime() + ']' + ' <'+ title + '>' + ': ' + '\n ' + context;  // append result to infoPrint1

    if (shouldScroll) {
      setTimeout(() => {
        infoPrintElement1.scrollTop = infoPrintElement1.scrollHeight;
      }, 0);
    }
  }

  function handleResult2(title, context)  {
    if (!isPrinting) return; // add this line to check the printing state

    if (selectedClient !== cliDisPlaySelectAll && selectedClient !== title) {
      return
    }

    let lineFeed = ""

    if (infoPrint2 !== '') {
      lineFeed = '\n'
    }
    let shouldScroll = infoPrintElement2.scrollTop + infoPrintElement2.clientHeight === infoPrintElement2.scrollHeight;

    if (isHex){
      //context转化为16进制显示
        let hexStr = ''
        for (let i = 0; i < context.length; i++) {
          hexStr += context.charCodeAt(i).toString(16) + ' '
        }
        context = hexStr
    }
    infoPrint2 += lineFeed + '[' + getCurrentDatetime() + ']' + ' <'+ title + '>' + ': ' + '\n ' + context;  // append result to infoPrint1

    if (shouldScroll) {
      setTimeout(() => {
        infoPrintElement2.scrollTop = infoPrintElement2.scrollHeight;
      }, 0);
    }
  }

  function showServerMsg(){
    ReadRemoteMsg().then(result => {
        if (result !== '') {
          let jsonObject = JSON.parse(String(result));
          handleResult1(jsonObject.ipAddr, jsonObject.msg);
        }
    })
  }

  function showClientsMsg(){
    ReadClientsMsg().then(result => {
        if (result !== '') {
          let jsonObject = JSON.parse(String(result));
          handleResult2(jsonObject.ipAddr, jsonObject.msg);
        }
    })
  }

  function checkAppStop(){
    CheckAppStop().then(result => {
      if (result === true) {
        AppTaskClose()
        showModal("异常终止,可能目标端断开连接")
      }
    })
  }

  function updateClientSelect(){
    GetClientsList().then(result => {
      if (result !== '') {
        let jsonObj = JSON.parse(String(result))
        let options = []
        if (jsonObj.clientsList !== null) {
          for (let i = 0; i < jsonObj.clientsList.length; i++) {
            options.push(jsonObj.clientsList[i])
          }
          //排序
          options.sort(function(a,b){
            return a.localeCompare(b)
          })
        }
        updateClientOptions(options)
      }
    })
  }

  function AppTaskClose() {
    clearInterval(checkAppStopTask)
    clearInterval(showServerMsgTask)
    clearInterval(showClientMsgTask)
    document.getElementById("start_stop").innerHTML = "Start"
    updateClientOptions([])
    startOrStop = false
    document.getElementById("start_stop").style.backgroundColor = "#007bff";
  }

  // handle the start/stop button click event
  function handleStartStop() {
    if (!startOrStop) {
       AppStart(targetIp, targetPort, listenIp, listenPort).then(result => {
         if (result === 'success') {
           document.getElementById("start_stop").innerHTML = "Stop"
           document.getElementById("start_stop").style.backgroundColor = "red";
           startOrStop = true
           showServerMsgTask = setInterval(showServerMsg, 5)
           showClientMsgTask = setInterval(showClientsMsg, 5)
           checkAppStopTask = setInterval(checkAppStop, 100)
           updateClientSelectTask = setInterval(updateClientSelect, 100)
         } else {
           showModal(result)
           console.log(result)
         }
       })
    } else {
      AppTaskClose()
      AppStop()
    }
  }

  // handle the client selection change event
  function handleClientChange(e) {
    // your logic here
    selectedClient = e.target.value
  }

  // a function to start or stop the printing
  function togglePrint() {
    isPrinting = !isPrinting; // toggle the printing state
    // change the button text
    if (isPrinting) {
      document.getElementById("toggle_print").innerHTML = "Stop Printing";
    } else {
      document.getElementById("toggle_print").innerHTML = "Start Printing";
    }
  }

  function asciiOrHex() {
    isHex = !isHex; // toggle the printing state
    // change the button text
    if (isHex) {
      document.getElementById("ascii_hex").innerHTML = "Hex";
    } else {
      document.getElementById("ascii_hex").innerHTML = "ASCII";
    }
  }

  // add a function to clear the log
  function clearLog() {
    infoPrint1 = '';
    infoPrint2 = '';
  }

</script>

<main>
  <div class="container">
    <div class="left">
      <hr>
      <span class="a-verdana">Target Ip</span>
      <input bind:value={targetIp} type="text" id="target_id" placeholder="Target IP">
      <span class="a-verdana">Target Port</span>
      <input bind:value={targetPort} type="text" id="target_port" placeholder="Target Port">
      <hr>
      <span class="a-verdana">Listen Ip</span>
      <input bind:value={listenIp} type="text" id="listen_ip" placeholder="Listen IP">
      <span class="a-verdana">Listen Port</span>
      <input bind:value={listenPort} type="text" id="listen_port" placeholder="Listen Port">
      <hr>
      <span class="a-verdana">Client Msg Display ({clientCount})</span>
      <select bind:value={selectedClient} on:change={handleClientChange} id="client_select" size="1">
        <option value="ALL" selected>ALL</option>
        <!-- Add other options dynamically -->
        {#each clientOptions as option}
          <option value={option}>{option}</option>
        {/each}
      </select>
      <hr>
      <button on:click={handleStartStop} id="start_stop">Start</button>
    </div>
    <div class="right">
      <label for="info_print1">Server Msg</label>
      <textarea id="info_print1" bind:this={infoPrintElement1} readonly bind:value={infoPrint1}></textarea>
      <label for="info_print2">Clients Msg</label>
      <textarea id="info_print2" bind:this={infoPrintElement2} readonly bind:value={infoPrint2}></textarea>
      <button id="toggle_print" on:click={togglePrint}>Stop Print</button>
      <button on:click={clearLog}>Clear Log</button>
      <button id="ascii_hex" on:click={asciiOrHex}>ASCII</button>
    </div>
  </div>
  {#if isModalOpen}
    <div class="modal">
      <div class="modal-content">
        <h2>错误信息</h2>
        <p>{modalContent}</p> <!-- 显示 modalContent 的内容 -->
        <button on:click={() => isModalOpen = false}>Close</button>
      </div>
    </div>
  {/if}
</main>



<style>
  .container {
    width: 100%;
    height: 100vh;
    display: flex;
    background-color: #222;
    color: #fff;
    font-family: Arial, sans-serif;
  }
  .left {
    width: 25%;
    padding: 10px;
    box-sizing: border-box;
  }
  .right {
    width: 75%;
    padding: 10px;
    box-sizing: border-box;
  }
  .left input, .left button {
    display: block;
    margin-bottom: 10px;
    width: 100%;
    height: 30px;
    border: none;
    border-radius: 4px;
    padding: 5px;
    background-color: #333;
    color: #fff;
  }
  .left input::placeholder {
    color: #999;
  }
  .left select {
    display: block;
    margin-bottom: 10px;
    height: 30px;
    width: 100%;
    border: none;
    border-radius: 4px;
    padding: 5px;
    background-color: #333;
    color: #fff;
  }
  .right textarea {
    width: 100%;
    height: calc(50% - 40px); /* subtract the height of the two buttons */
    box-sizing: border-box;
    border: none;
    border-radius: 4px;
    padding: 5px;
    background-color: #333;
    color: #fff;
    resize: none;
  }
  .right label {
    display: block;
    margin-bottom: 5px;
    color: #fff;
  }
  body, html {
    margin: 0;
    padding: 0;
    height: 100%;
    box-sizing: border-box;
    background-color: #111;
  }
  .a-verdana {
    font-family:  Georgia, serif;
  }
  #start_stop {
    width: 100%; /* the width of the button */
    height: 30px; /* the height of the button */
    background-color: #007bff;
    color: #fff;
    cursor: pointer;
  }
  #start_stop:hover {
    background-color: #0056b3;
  }
  button {
    width: 100px; /* the width of the button */
    height: 30px; /* the height of the button */
    border: none;
    border-radius: 4px;
    padding: 5px 10px;
    margin-right: 10px;
    background-color: #007bff;
    color: #fff;
    cursor: pointer;
  }
  button:hover {
    background-color: #0056b3;
  }

  /* Modal styles */
  .modal {
    position: fixed;
    z-index: 1;
    left: 0;
    top: 0;
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: rgba(0, 0, 0, 0.4);
  }

  .modal-content {
    background-color: #333;
    color: #fff;
    padding: 20px;
    border: 1px solid #888;
    width: 80%;
    max-width: 400px; /* Adjust the width as needed */
    text-align: center;
    border-radius: 4px;
  }

  .modal-content h2 {
    margin-top: 0;
  }
  .modal-content p {
    margin-bottom: 20px;
  }
  .modal-content button {
    padding: 5px 10px;
    background-color: #007bff;
    color: #fff;
    border: none;
    border-radius: 4px;
    cursor: pointer;
  }
  .modal-content button:hover {
    background-color: #0056b3;
  }
  hr {
    border-color: #464646; /* 灰色 */
  }

</style>
