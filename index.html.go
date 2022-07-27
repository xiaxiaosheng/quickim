/*
 * Copyright 2019 the go-netty project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

var indexHtml = []byte(`
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8" /> <meta name="viewport" content="target-densitydpi=device-dpi, width=device-width, initial-scale=1, user-scalable=no, minimum-scale=1.0, maximum-scale=1.0"/>
<title>ChatRoom</title>
</head>
<body>
<style>
body{
      background-color: #ebebeb;
      font-family: -apple-system;
      font-family: "-apple-system", "Helvetica Neue", "Roboto", "Segoe UI", sans-serif;
    }
    .chat-sender{
      clear:both;
      font-size: 80%;
    }
    .chat-sender div:nth-of-type(1){
      float: left;
    }
    .chat-sender div:nth-of-type(2){
      margin: 0 50px 2px 50px;
      padding: 0px;
      color: #848484;
      font-size: 70%;
      text-align: left;
    }
    .chat-sender div:nth-of-type(3){
      background-color: white;
      /*float: left;*/
      margin: 0 50px 10px 50px;
      padding: 10px 10px 10px 10px;
      border-radius:7px;
      text-indent: -12px;
    }

    .chat-receiver{
      clear:both;
      font-size: 80%;
    }
    .chat-receiver div:nth-of-type(1){
      float: right;
    }
    .chat-receiver div:nth-of-type(2){
      margin: 0px 50px 2px 50px;
      padding: 0px;
      color: #848484;
      font-size: 70%;
      text-align: right;
    }
    .chat-receiver div:nth-of-type(3){
      /*float:right;*/
      background-color: #b2e281;
      margin: 0px 50px 10px 50px;
      padding: 10px 10px 10px 10px;
      border-radius:7px;
    }

    .chat-receiver div:first-child img,
    .chat-sender div:first-child img{
      width: 40px;
      height: 40px;
      /*border-radius: 10%;*/
    }

    .chat-left_triangle{
      height: 0px;
      width: 0px;
      border-width: 6px;
      border-style: solid;
      border-color: transparent white transparent transparent;
      position: relative;
      left: -22px;
      top: 3px;
    }
    .chat-right_triangle{
      height: 0px;
      width: 0px;
      border-width: 6px;
      border-style: solid;
      border-color: transparent transparent transparent #b2e281;
      position: relative;
      right:-22px;
      top:3px;
    }

    .chat-notice{
      clear: both;
      font-size: 70%;
      color: white;
      text-align: center;
      margin-top: 15px;
      margin-bottom: 15px;
    }
    .chat-notice span{
      background-color: #cecece;
      line-height: 25px;
      border-radius: 5px;
      padding: 5px 10px;
    }
</style>
		<form id='main_frm' align="left" style="width: 100%" onsubmit="return false;">
			<div>
				<h3>
					<font>房间号:</font>
					<font id="roomId">------</font>
				</h3>
				<div >
					<input type="text" id="changeRoomId" name="changeRoomId" style="width: 100px;" value="------">
					<input  type="button" value="切换房间" onclick="changeRoom(this.form.changeRoomId.value)">
					<input  type="button" value="分享房间" onclick="copyShareLink()">
				</div>
				<div >
					<input type="text" id="nickname" name="nickname" style="width: 100px;" value="------">
					<input  type="button" value="切换昵称" onclick="changeNickname(this.form.nickname.value)">
					<input  type="button" value="清空消息" onclick="clearRoomMessage()">
				</div>
	
			</div>
			<div id="responseText" style="width: 97%;height: 300px; overflow:scroll; background-color: #F0F0F0; margin: 0.5%;">
			</div>
			<textarea id="editText" name="editText" style="width: 95%; height: 50px; margin-top:5px;"></textarea>
			<input id="send_btn" class="btn"  style="float:right; margin-right:2.5%;" type="button" value="发送" onclick="send(this.form.editText.value)">
			<!-- <input type="button" onclick="javascript:document.getElementById('responseText').value=''" value="Clear"> -->
		</form>
		</br>
		<font style="font-weight:bold;">在线用户：<font id="totalOnline">0</font></font>
		<ul id="onlineList">
		</ul>
</body>
	<script src="https://cdn.bootcss.com/jsencrypt/3.0.0-beta.1/jsencrypt.js"></script>
	<script src="https://cdn.bootcdn.net/ajax/libs/jquery/2.2.0/jquery.js" type="text/javascript" charset="utf-8"></script>
	<script src="https://cdn.bootcss.com/crypto-js/3.1.9-1/crypto-js.min.js"></script>
	<script src="/coder.js"></script>
	<script type="text/javascript">
		var onlineList = [];
		var totalOnline = 0;
		var otherColor = 'blue'; // 通知字体颜色
		var leftAlign = 'left'; // 左对齐
		var rightAlign = 'right'; // 右对齐
		var desKey = ''; // 密钥


		function rsaEncode(str, publicKey) {
			var encrypt = new JSEncrypt();
			encrypt.setPublicKey(publicKey);
			//encrypt.setPublicKey('-----BEGIN PUBLIC KEY-----' + publicKey + '-----END PUBLIC KEY-----');
			
			var cipher = encrypt.encrypt(str);
			return cipher
		}
		
		//DES加密
		function encryptByDES(message, key){
		    var keyHex = CryptoJS.enc.Utf8.parse(key);
		    var encrypted = CryptoJS.DES.encrypt(message, keyHex, {
		        mode: CryptoJS.mode.ECB,
		        padding: CryptoJS.pad.Pkcs7
		    });
		    return encrypted.ciphertext.toString();
		}
		
		//DES解密
		function decryptByDES(ciphertext, key){
		    var keyHex = CryptoJS.enc.Utf8.parse(key);
		    var decrypted = CryptoJS.DES.decrypt({
		        ciphertext: CryptoJS.enc.Hex.parse(ciphertext)
		    }, keyHex, {
		        mode: CryptoJS.mode.ECB,
		        padding: CryptoJS.pad.Pkcs7
		    });
		    var result_value = decrypted.toString(CryptoJS.enc.Utf8);
		    return result_value;
		}

		function setHTML() {
			var tmp = "";
			if(onlineList) {
				for(var i = 0; i < onlineList.length; i++) {
					tmp += ('<li>'+onlineList[i] + '</li>');
				}
			}
			document.getElementById('onlineList').innerHTML = tmp;
			document.getElementById('totalOnline').innerHTML = totalOnline.toString();
		}
		function setOnline(arr) {
			onlineList = arr
			if(arr) {
				totalOnline = arr.length;
			}
			setHTML();
		}
		function addOnline(name) {
			onlineList[onlineList.length] = name;
			totalOnline++;
			setHTML();
		}
		function subOnline(name) {
			var tmp = [];
			var f = 0;
			for(var i = 0; i < onlineList.length; i++) {
				if(name != onlineList[i]) {
					tmp[f++] = onlineList[i];
				}
			}
			onlineList = tmp;
			totalOnline--;
			setHTML();
		}
		function getQueryVariable(variable) {
		       var query = window.location.search.substring(1);
		       var vars = query.split("&");
		       for (var i=0;i<vars.length;i++) {
		               var pair = vars[i].split("=");
		               if(pair[0] == variable){return pair[1];}
		       }
		       return(false);
		}
		var meHeadImg = 'data:image/jpeg;base64,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';
		var otherHeadImg = 'data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEAYABgAAD/4QBORXhpZgAATU0AKgAAAAgAAgESAAMAAAABAAEAAIdpAAQAAAABAAAAJgAAAAAAAqACAAQAAAABAAABeqADAAQAAAABAAABmQAAAAAAAP/iAiBJQ0NfUFJPRklMRQABAQAAAhBhcHBsBAAAAG1udHJSR0IgWFlaIAfmAAcAGgALABkAGmFjc3BBUFBMAAAAAEFQUEwAAAAAAAAAAAAAAAAAAAAAAAD21gABAAAAANMtYXBwbP4w5aAK84VzKnbbquB2lJIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACmRlc2MAAAD8AAAAYWNwcnQAAAFgAAAAI3d0cHQAAAGEAAAAFHJYWVoAAAGYAAAAFGdYWVoAAAGsAAAAFGJYWVoAAAHAAAAAFHJUUkMAAAHUAAAAEGNoYWQAAAHkAAAALGJUUkMAAAHUAAAAEGdUUkMAAAHUAAAAEGRlc2MAAAAAAAAABzI0OTBXMQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAB0ZXh0AAAAAENvcHlyaWdodCBBcHBsZSBJbmMuLCAyMDIyAABYWVogAAAAAAAA89gAAQAAAAEWCFhZWiAAAAAAAABk5QAAM20AAAEoWFlaIAAAAAAAAGoQAAC79gAAEbVYWVogAAAAAAAAJ+EAABCdAADAUHBhcmEAAAAAAAAAAAAB9gRzZjMyAAAAAAABC7cAAAWW///zVwAABykAAP3X///7t////aYAAAPaAADA9v/bAEMAAgEBAgEBAgICAgICAgIDBQMDAwMDBgQEAwUHBgcHBwYHBwgJCwkICAoIBwcKDQoKCwwMDAwHCQ4PDQwOCwwMDP/bAEMBAgICAwMDBgMDBgwIBwgMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDP/AABEIAEUAQAMBIgACEQEDEQH/xAAfAAABBQEBAQEBAQAAAAAAAAAAAQIDBAUGBwgJCgv/xAC1EAACAQMDAgQDBQUEBAAAAX0BAgMABBEFEiExQQYTUWEHInEUMoGRoQgjQrHBFVLR8CQzYnKCCQoWFxgZGiUmJygpKjQ1Njc4OTpDREVGR0hJSlNUVVZXWFlaY2RlZmdoaWpzdHV2d3h5eoOEhYaHiImKkpOUlZaXmJmaoqOkpaanqKmqsrO0tba3uLm6wsPExcbHyMnK0tPU1dbX2Nna4eLj5OXm5+jp6vHy8/T19vf4+fr/xAAfAQADAQEBAQEBAQEBAAAAAAAAAQIDBAUGBwgJCgv/xAC1EQACAQIEBAMEBwUEBAABAncAAQIDEQQFITEGEkFRB2FxEyIygQgUQpGhscEJIzNS8BVictEKFiQ04SXxFxgZGiYnKCkqNTY3ODk6Q0RFRkdISUpTVFVWV1hZWmNkZWZnaGlqc3R1dnd4eXqCg4SFhoeIiYqSk5SVlpeYmZqio6Slpqeoqaqys7S1tre4ubrCw8TFxsfIycrS09TV1tfY2dri4+Tl5ufo6ery8/T19vf4+fr/2gAMAwEAAhEDEQA/AP38oooZtq5PAHJJ7UAFFVxq9q3S6tz/ANtBR/a9r/z9W/8A38FAHm/wj+OF146/aE+K/g26WER+B7rTTZsi7WaC6skkIbn5iJVl544ZR2r1Cvk/9mPXTff8FL/jske1oZ7SxJYHqYYoYx/6E1fWFcuEqOcG33kvubOXCVHODb7yX3NhRRRXUdQVDqNmuo2E1uzMqzI0ZI6gEYrmfjX8YdL+BPw5vvEmrLcTW9qUiitrZN9xezyMEihiX+J3dgB2HJJABI+XLr/gqzrngrWmj8YfCnVPD9pcQyPaK9zKlyzBTsG2aGMOpfaGZSNoJIDYAPJiMdQoyUKkrNnPWxdKk7VHb5P8exd/aw8IfCD4B2f/ABVF3rmqazqgkuLfR7LULj7RcB/NVpGHnKkUZMsvzMQCd20MygD5Q8SftG3F9ceXpfhnQ9P0uF3MVtdXuo3Uw3SCU7pVuYg3zKDjYAOQOCapeF/Dni/9rf43NCsi6p4o8STNdXl1MSsMCLgPIx5KQRrtRVGcDy0XJKg/eHwZ/YC+Hfwo02H7bpNv4u1cAGW+1iBZ4y3H+rt2zFGoPThnHd2614UsdisTK9J8sf6/rseXRq47HSboPlguv9dT4Z+Ff7Vfi/4S/FbxB4w0i7019a8TPnURcWnm28g3bwioGDKo4Aw2cAcnrX1d8CP+Cs2ma/qEOn/EDSI9BaQhBq1gzy2Sn1ljbMkS/wC0GkA6ttAJr3bxR8AfAvjPT/suqeDvDN3Dt2qTp0SSRD/YdVDIfdSDXxz+2b+wHH8HtAufF/gyS6uPDtmN+oadcSGWbS0yB5sbn5pIV43B8ugG4sy7imcZYrDLmhK63a/P+kFTA47BRdSnLmju1+bt/kz9ENO1G31ewgurWaG5tbqNZYZonDxyowyrKw4KkEEEcEVNX5p/sjf8FAL/APZq+HuqeGLvSbvxNG0qP4ftvtHlLbSO2JYWfDFYySrIqIx3lx0YFfqL9nP9ufUfiX49svC/jbwLq3gXVdYVzpE9wk32bUnRC7xAyRIVcICwxuBCtyp2q3tYfNqFXli3aT6efqdmGzKlViujfTXR+u3od3+0/bLb6Z4R1e4VW0/QfEcNxdswysSzW9zZxyHthJrqFiTwoUtxtyPHf24PD7eIf2XPFSLCs0lgtvfruQMYRDcxSSuM/dIiWUEjsWHQmvqnWNItfEGk3Vhf21veWN9E9vcW88YkinjdSrIynhlIJBB4INeRav8AAnxN4ds5tN0i40fxX4euEeBbDxDcy29zBCwIMTXSRzfaIwp2gSRCTH35ZCc15+cZbVq1FWpK/Ro9inWgqNTD1Npp672bVv8AI+dv+CSOkWjXHj/UGCtfxiwtlb+KOJvtDnHs7IM+vlD0r6AvvFvxNh/ar0/R4fDOlyfC6TTWlutXMn+krcbHIA/eAg+YI02eWcqzPu7L8b/Br4oW37Bn7VutaXdX1nqnhmdjp2qnTZ5L0WAVy0eZGijMs1sSyuFXBV5MAyDYv0p4Z/b3+ENp49ttD0/UtUjW4kS0i1OW1mGnliQq5eRt4BYKDIUC5+YtjLVx4apGMFCTs09dfP8AE8XK8Th44ZUKlTllGXS2ur/Bnc/tK/ELxpo3wxkk+FOj6T4m8VNPat5Et1Gy29tJuYTFC6b1fZsHzrwxcEhGFd7Fpy+IdCW01S1t5F1C2EN7bq3mQuHTEiAn7ynLDPcVz/wx+A/hP4NWOoWfhjQ4NJh1R1NzHHJJJuVE2RxLvZjHDGmVjhTbHGCQiqCa8f8A2wP2oPD/AMDfCmoeFvCs1pP4816MabFbWTB20veWHmy84SQea5RG+YswYjYGI6pVJR9+p93T5aK9z3sdLD4fnqqcnGysmkvwTerfmz5V/YK8OnVP2m9DeHF1DpNpd3jTMoYlBCYFk9jvmj596+3PGVmuv+J/Bekx5bULnxDa3sBAyYEtW+0TyeymJGhLdM3KL/GAfOf2TP2LfHfwD8N3MjaX4P8A7a11YhcXFxrM7NpsKjKQiGO2KuyksX2zAM2AGIVTX0V8L/g3H4F1G51jUr5tc8S30Ygmv2hEEVvCDu+z20WW8mHd8xBZ3chS7vtTbx4LKa8q0ZVFypO55uU2w+XPDte9NtvyWit5uy6bM7avF/2+Pjrc/AT9nLUr7TZ2ttc1qVNH0yVfvQSyhi0oPZo4Uldc8bkXPFe0V8p/8FcvCV7rXwC0PVrZJZLbw/rkc97t+7DFJDNCJG9hI8a57eZnpX02OnKOHnKG9jPHTlHDzlDex+dscAlHIYRxnEak/ex/EfX2+mfTEl6FexmV1LK6EbSM5yD/ADzj8aecJGc/dUZOT2+tRRj7VMrcrCjBl3D/AFh9foM8ep56Dn4E+C6HX618cfHXiKyNrqHjbxlfWsi4eGbW7t45Bjoy+ZtI+orkktkCFVjUKeMBQBz7e/P507O8deW68dPpTZSxhbbjcASMj/P+TVSk3uXKcpfE7n6F/wDBL39p24+JPg268Ca5dNPrXheBZrCaRt0l3p+QgBPdoWKoSeqyRdTuNfV9fkX+yZ8T5PhL+0d4N1xZDHatqMdjdknCm2uSIZC3qF3iT6xr6V+ulfYZPinVoWlvHT5dD67J8U6tC0t46fLofMP/AA1R4g9F/wC+/wD61RX37TOsarZTWt1b211a3SNFNDMBJHMjDDKysCGUgkEEYINFFesesfLniv8AYsg8VeJtR1Oy14aPZ31y9xb6fDp4aCxRjkRoC4GF6DgdBxVU/sMXJHPjW4+Y/wDQNHH/AJEoorglluF5r8n5nG8rwr15PzGj9he5Ukr41n9h/Zi/r+8pT+wxcAf8jpcdP+gav/xyiil/ZeFv8H5/5h/ZOEt8H4v/ADIpP2Drg2Xkr44u0ITZvGnLuHHUfvOtfXDftUeIM/w/99f/AFqKK6MPhaVG/s1a5tSwtKh/CVrn/9k=';
		
		function appendMsg(t, name, txt) {
			var respTxt = document.getElementById('responseText');
			var txtHtml = '';

			if(t == 0) {
				txtHtml = '<div class="chat-receiver"><div><img src="' + meHeadImg + '"></div> <div>' + name + '</div> <div> <div class="chat-right_triangle"></div><span>' + txt + '</span> </div></div>';
			} else {
				txtHtml = '<div class="chat-sender"><div><img src="' + otherHeadImg +  '"></div><div>' + name + '</div><div><div class="chat-left_triangle"></div><span>'+ txt + '</span> </div></div>';
			}

			// var txtHtml = '<p align="' + align + '"><font color="' + color + '">'+txt + '</font>';
			respTxt.innerHTML += (txtHtml);
			respTxt.innerHTML += "<div id='msgEnd' style='height:0px; overflow:hidden;'></div>";
			var msgEnd = document.getElementById('msgEnd');
			msgEnd.scrollIntoView();
			respTxt.removeChild(msgEnd);
		}

		function clearMsg() {
			var respTxt = document.getElementById('responseText');
			respTxt.innerHTML = '';
		}

		function OnMessage(event) {
			var cmd = JSON.parse(decode(event.data));
			// console.log(cmd);
			switch(cmd.code) {
			case "RNM":
				document.getElementById('roomId').innerHTML = cmd.roomId
				document.getElementById('changeRoomId').value = cmd.roomId
				document.getElementById('nickname').value = cmd.nickname
				setOnline(cmd.onlineList)
				break;
			case "JNR":
				appendMsg(1, "系统", cmd.name + ' 加入了房间!');
				addOnline(cmd.name);
				break;
			case "OUT":
				appendMsg(1, '系统', cmd.name + ' 退出了房间!');
				subOnline(cmd.name);
				break;
			case "MSG":
				var t = 1;
				if(typeof(cmd.is_self) != "undefined" && cmd.is_self == 1) { // 说明是自己的历史消息
					t = 0;
				}
				appendMsg(t, cmd.from + '(' + cmd.sendTime + ')',  cmd.message);
				break;
			case "OLT": // 在线用户列表
				setOnline(cmd.onlineList);
				break;
			case "SHAKESUCC": // 握手成功
				login();
				break;
			case "PUBKEY": // 下发公钥
				shake(cmd);
				break;
			}
			return;
		}

		function copyShareLink() {
			const input = document.createElement('input');
			document.body.appendChild(input);
			input.setAttribute('value', document.location.origin+'?roomId=' + document.getElementById('changeRoomId').value);
			input.select();
			if (document.execCommand('copy')) {
			        document.execCommand('copy');
			        alert('已复制房间分享链接，去粘贴吧！');
			}
			 document.body.removeChild(input);
		}

		function uuid() {
		    var s = [];
		    var hexDigits = "0123456789abcdef";
		    for (var i = 0; i < 36; i++) {
		        s[i] = hexDigits.substr(Math.floor(Math.random() * 0x10), 1);
		    }
		    s[14] = "4"; // bits 12-15 of the time_hi_and_version field to 0010
		    s[19] = hexDigits.substr((s[19] & 0x3) | 0x8, 1); // bits 6-7 of the clock_seq_hi_and_reserved to 01
		    s[8] = s[13] = s[18] = s[23] = "-";
		
		    var uuid = s.join("");
		    return uuid;
		}
		var Cookie = {
		    set: function (key, value, exdays) {
		        let exdate = new Date() // 获取时间
		        exdate.setTime(exdate.getTime() + 24 * 60 * 60 * 1000 * exdays) // 保存的天数
		        // 字符串拼接cookie
		        // eslint-disable-next-line camelcase
		        window.document.cookie = key + '=' + value + ';path=/;expires=' + exdate.toGMTString()
		    },
		
		    get: function (key) {
		        if (document.cookie.length > 0) {
		            var arr = document.cookie.split('; ') // 这里显示的格式需要切割一下自己可输出看下
		            for (let i = 0; i < arr.length; i++) {
		                let arr2 = arr[i].split('=') // 再次切割
		                // 判断查找相对应的值
		                if (arr2[0] === key) {
		                    return arr2[1]
		                }
		            }
		        }
		    },
		
		    remove: function (key) {
		        set(key, '', -1);
		    }
		};


		var socket;
		if (!window.WebSocket) {
			window.WebSocket = window.MozWebSocket;
		}
		function write(cmd) {
			if (!window.WebSocket) {
				return;
			}
			if (socket.readyState == WebSocket.OPEN) {
				var body = encode(JSON.stringify(cmd));

				socket.send(body);
			} else {
				document.getElementById('roomId').innerHTML = "------";
				document.getElementById('changeRoomId').value = "------";
				document.getElementById('nickname').value = "------";
				onlineList = [];
				setHTML();
			}
		}
		function login() {
			var token = Cookie.get("token");
			if(!token) {
				token = uuid();
				Cookie.set("token", token, 100)
			}
			// console.log(token);
			write({"code": "LOGIN", "user_id": token})

			var roomId = getQueryVariable('roomId');
			if(roomId != "") {
				changeRoom(roomId);
			}
		}

		function shake(cmd) {
			var des = uuid();
			var code = rsaEncode(des, cmd.public_key);
			write({"code": "KEY", "key": code});
			desKey = des; // 进入des加密通信
		}

		if (window.WebSocket) {
			socket = new WebSocket("ws://" + window.location.host + "/chat");
			socket.onmessage = OnMessage;
			socket.onopen = function(event) {
				// 启动心跳发送
				setTimeout(live, 3000)

				window.onbeforeunload = function () {
					out()
				};
			};
			socket.onclose = function(event) {
			};
			
		} else {
			alert("Your browser does not support WebSocket!");
		}
		

		function live() { // 发送心跳包
			write({"code": "LIV"})
			setTimeout(live, 3000)
		}
		function out() {
			write({"code": "OUT"})
		}
		function clearRoomMessage() {
			clearMsg();
			write({"code": "CLEAR"})
		}
		function send(message) {
			if(message == "") return;
			write({"code": "MSG", "message": message})
			var t = new Date().toLocaleTimeString('chinese', { hour12: false });
			var title = document.getElementById('nickname').value + '(' + t + ')';
			appendMsg(0, title, message);
			document.getElementById('editText').value = "";
		}
		function changeRoom(roomId) {
			clearMsg();
			write({"code": "CHM", "roomId": roomId})
		}
		function changeNickname(name) {
			write({"code": "CHN", "name": name})
		}

		function stringToByte(str) {
	            var bytes = new Array();
	            var len, c;
	            len = str.length;
	            for(var i = 0; i < len; i++) {
	                c = str.charCodeAt(i);
	                if(c >= 0x010000 && c <= 0x10FFFF) {
	                    bytes.push(((c >> 18) & 0x07) | 0xF0);
	                    bytes.push(((c >> 12) & 0x3F) | 0x80);
	                    bytes.push(((c >> 6) & 0x3F) | 0x80);
	                    bytes.push((c & 0x3F) | 0x80);
	                } else if(c >= 0x000800 && c <= 0x00FFFF) {
	                    bytes.push(((c >> 12) & 0x0F) | 0xE0);
	                    bytes.push(((c >> 6) & 0x3F) | 0x80);
	                    bytes.push((c & 0x3F) | 0x80);
	                } else if(c >= 0x000080 && c <= 0x0007FF) {
	                    bytes.push(((c >> 6) & 0x1F) | 0xC0);
	                    bytes.push((c & 0x3F) | 0x80);
	                } else {
	                    bytes.push(c & 0xFF);
	                }
	            }
	            return bytes;
	        }
		 function byteToString(arr) {
	            if(typeof arr === 'string') {
	                return arr;
	            }
	            var str = '',
	                _arr = arr;
	            for(var i = 0; i < _arr.length; i++) {
	                var one = _arr[i].toString(2),
	                    v = one.match(/^1+?(?=0)/);
	                if(v && one.length == 8) {
	                    var bytesLength = v[0].length;
	                    var store = _arr[i].toString(2).slice(7 - bytesLength);
	                    for(var st = 1; st < bytesLength; st++) {
	                        store += _arr[st + i].toString(2).slice(2);
	                    }
	                    str += String.fromCharCode(parseInt(store, 2));
	                    i += bytesLength - 1;
	                } else {
	                    str += String.fromCharCode(_arr[i]);
	                }
	            }
	            return str;
	        }

		function numToChar(b) {
			if(b >= 0 && b <= 9) {
				return String.fromCharCode(b + 48)
			} else {
				return String.fromCharCode(b + 87)
			}
		}
		function charToNum(b) {
			if(b >= 48 && b <= 57) {
				return b - 48
			} else if(b >= 97 && b <= 122) {
				return b - 87
			} else {
				return b - 55
			}
		}
		function encode(str) {
			if(desKey == "") {
				return str;
			} else {
				return encryptByDES(str, desKey);
			}
		}
		function decode(str) {
			if(desKey == "") {
				return str;
			} else {
				return decryptByDES(str, desKey);
			}
		}

		function resizeScreen() {
			var w = $(document).width();
			var h = $(document).height();
			if(w > 0.8*h) {
				$('#main_frm').attr('style', "width: 45%");

			}
		}
		resizeScreen();

		$(document).keydown(function(e){
		    if(e.keyCode == 13){
		    	$("#send_btn").click();
				return false;
		    }
		});


	</script>

</html>
`)

var namePool = []string{"亚托克斯", "布里茨", "德莱厄斯", "黛安娜", "蒙多医生", "艾克", "伊莉丝", "菲奥娜", "菲兹", "普朗克", "纳尔",
	"古拉加斯", "盖伦", "赫卡里姆", "俄洛伊", "艾瑞莉娅", "嘉文四世", "贾克斯", "杰斯", "凯尔", "卡兹克", "李青", "墨菲特", "易",
	"孙悟空", "莫德凯撒", "内瑟斯", "诺提勒斯", "奈德丽", "魔腾", "努努", "奥拉夫", "潘森", "波比", "奎因", "拉莫斯", "雷克塞", "雷克顿",
	"雷恩加尔", "锐雯", "兰博", "瑞兹", "瑟庄妮", "希瓦娜", "辛吉德", "赛恩", "斯卡纳", "斯维因", "泰隆", "塔里克", "锤石", "特朗德尔",
	"泰达米尔", "乌迪尔", "厄加特", "蔚", "沃利贝尔", "沃里克", "赵信", "亚索", "约里克", "扎克", "劫", "九尾妖狐", "殇之木乃伊",
	"冰晶凤凰", "黑暗之女", "铸星龙王", "沙漠皇帝", "星界游神", "复仇焰魂", "魔蛇之拥", "虚空恐惧", "皎月女神", "蜘蛛女皇", "寡妇制造者",
	"探险家", "末日使者", "哨兵之殇", "酒桶", "大发明家", "风暴之怒", "天启者", "死亡颂唱者", "虚空行者", "不祥之刃", "狂暴之心", "深渊巨口",
	"诡术妖姬", "冰霜女巫", "仙灵女巫", "光辉女郎", "虚空先知", "扭曲树精", "堕落天使", "唤潮鲛姬", "发条魔灵", "机械公敌", "符文法师",
	"琴瑟仙女", "众星之子", "策士统领", "暗黑元首", "岩雀", "卡牌大师", "惩戒之箭", "邪恶小法师", "虚空之眼", "机械先驱", "猩红收割者",
	"远古巫灵", "掘墓者", "爆破鬼才", "时光守护者", "荆棘之兴", "暗裔剑魔", "牛头酋长", "殇之木乃伊", "蒸汽机器人", "弗雷尔卓德之心",
	"虚空恐惧", "诺克萨斯之手", "祖安狂人", "哨兵之殇", "德玛西亚之力", "迷失之牙", "战争之影", "海兽祭司", "德玛西亚皇子", "曙光女神",
	"熔岩巨兽", "扭曲树精", "齐天大圣", "沙漠死神", "深海泰坦", "狂战士", "圣锤之毅", "披甲龙龟", "荒漠屠夫", "凛冬之怒", "暮光之眼",
	"龙血武姬", "炼金术士", "亡灵战神", "水晶先锋", "河流之王", "巨魔之王", "兽灵行者", "猩红收割者", "雷霆咆哮", "嗜血猎手", "生化魔人",
	"寒冰射手", "沙漠皇帝", "皮城女警", "英勇投弹手", "荣耀行刑官", "探险家", "法外狂徒", "未来守护者", "戏命师", "暴走萝莉", "复仇之矛",
	"狂暴之心", "永猎双子", "深渊巨口", "圣枪游侠", "赏金猎人", "德玛西亚之翼", "战争女神", "迅捷斥候", "麦林炮手", "瘟疫之源", "首领之傲",
	"惩戒之箭", "暗夜猎手", "牛头酋长", "冰晶凤凰", "寒冰射手", "星界游神", "弗雷尔卓德之心", "末日使者", "大发明家", "风暴之怒", "天启者",
	"审判天使", "曙光女神", "仙灵女巫", "光辉女郎", "堕落天使", "唤潮鲛姬", "雪人骑士", "发条魔灵", "琴瑟仙女", "众星之子", "暗黑元首",
	"河流之王", "岩雀", "瓦洛兰之盾", "魂锁典狱长", "时光守护者", "荆棘之兴"}
