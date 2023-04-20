const sendmsg = async function(room, msg) {
  data = {
    "Room": room,
    "Msg" : msg,
  }
  await fetch('http://localhost:8080/s/', {
    method: 'POST', 
    mode: 'no-cors',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  }
  )
}

let urlParams = new URLSearchParams(window.location.search);
const send = function () {
	let room = urlParams.get("room");
  var msg = document.getElementById('msg');
	sendmsg(room, msg.value);
}

const subscribeToRoom = function() {
	let room = urlParams.get("room");
	const es = new EventSource('http://localhost:8080/n/' + room);

	const listener = function (event) {
		var data = document.getElementById('data');
		console.log(event)
		data.innerText += event.data;
	}

	es.addEventListener('message', listener)
}

window.onload = function() {
	subscribeToRoom();
}
