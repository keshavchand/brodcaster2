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

const send = function () {
  var room = document.getElementById('Room');
  var msg = document.getElementById('Room');
}
