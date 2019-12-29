// vim: sw=2 ts=2 expandtab
(function () {
  "use strict";

  function addChange(changeType,direction,fieldName){
    fetch('/addChange',{
      method: 'post',
      headers: {
        "Content-type": "application/x-www-form-urlencoded; charset=UTF-8"
      },
      body: 'changeType='+encodeURIComponent(changeType) + "&direction="+encodeURIComponent(direction)+"&fieldName="+encodeURIComponent(fieldName)
    }).then(
      function(response) {
        if (response.status !== 200) {
          alert("Fetch Failed");
          document.title = 'fetch() Failed :(';
          console.log('fetch failedStatus Code: ' + response.status);
          return;
        }

        // Examine the text in the response
        response.json().then(function(data) {
          document.title = data;
          //console.log(data);
        });
      }
    ).catch(function(err) {
      document.title = 'fetch() Failed :-( ';
      console.log('Fetch Error :-S', err);
    });

  }
  window.addChange = addChange;
}());
