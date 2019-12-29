// vim: sw=2 ts=2 expandtab
(function () {
  "use strict";
  var pastId = null;
  function addChange(changeType,direction,fieldName,id,event,linkTargetObject){
    if(pastId!=null && event.altKey == true){
      document.title = 'SELECTED:' + pastId + " - " + id;
      id = pastId+","+id;
      pastId = null;
    }else if(pastId==null && event.altKey == true){
      pastId = id;
      document.title = 'SELECTED:' + pastId + " - ? ";
      return;
    }

    event.target.parentNode.style.color='red';
    fetch('/addChange',{
      method: 'post',
      headers: {
        "Content-type": "application/x-www-form-urlencoded; charset=UTF-8"
      },
      body: 'changeType='+encodeURIComponent(changeType) + "&direction="+encodeURIComponent(direction)+"&fieldName="+encodeURIComponent(fieldName)+"&id="+encodeURIComponent(id)
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
          if("result" in data){
            document.title = data.result;
          }else{
            document.title = JSON.Stringify(data);
          }
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
