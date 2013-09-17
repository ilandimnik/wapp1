(function() {
  $( document ).ready(function() {
    var websocket = new WebSocket("ws://localhost:3000/ws");

    websocket.onmessage = function(evt) {
      console.log("on message: Received event:" + evt.data )  
    };
    websocket.onclose= function(evt) {
      console.log("on close: Received event:" + evt.data)
    }
    websocket.onopen= function(evt) {
      console.log("on open: Received event:" + evt.data)
      // If we have session info, notify server we're ready
      if ($(".session-info").data("session")) {
        websocket.send(JSON.stringify({
          cmd: "connect",
          id: $(".session-info").data("session") 
        }))
      }
    }

  })
}());
