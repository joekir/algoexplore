const counterColour = "chartreuse";
const hitColours = ["red","dodgerblue","indigo"];

const elemFontSize = "4px",
     titleFontSize = "5px";

var cubeWidth = 10,
      xBuffer = 280,
      yBuffer = 40,
       svgDoc = d3.selectAll("svg");

var updateSizing = function(){
  if (typeof inputBytes !== 'undefined'){
    cubeWidth = xBuffer / inputBytes.length;
  }
};

let strToByteArr = function(str){
  var arr = [];
  for (var i = 0; i < str.length; i++) {
      arr.push(str.charCodeAt(i));
  }

  return arr;
}

let bitArray = function(arr) {
    let output = []; // there is no bit array :(
    for(let i=0; i < arr.length; i++){
      for (let j=0; j < 32; j++) {
          let mask = 1 << j;
          if ((arr[i] & mask) == mask) {
              output.push(1);
          } else {
              output.push(0);
          }
      }
    }

    return output;
}

let appendArray = function(title, backingArray, highlight, yIncrement){
  var items = svgDoc.selectAll("g");

  items.data([title])
       .enter()
       .append("text")
       .style("font-size", titleFontSize)
       .attr("x", xBuffer - title.length*0.40*cubeWidth)
       .attr("y", yBuffer)
       .text(d => d);

  items.data(backingArray)
       .enter()
       .append("rect")
       .attr("x", (d,i) => { return (xBuffer - i*cubeWidth) })
       .attr("y", yBuffer+0.4*cubeWidth)
       .attr("width", cubeWidth)
       .attr("height", cubeWidth)
       .style("fill", highlight);

  items.data(backingArray)
       .enter()
       .append("text")
       .text((d) => d.toString(16)) // mostly this will be bits, but if not hex it
       .style("font-size", elemFontSize)
       .attr("x", (d,i) => { return (xBuffer - i*cubeWidth + cubeWidth/5)})
       .attr("y", yBuffer + cubeWidth);

  yBuffer+=yIncrement*cubeWidth;
}

let appendText = function(titles, numbers){
  var items = svgDoc.selectAll("g");

  items.data(titles)
       .enter()
       .append("text")
       .style("font-size", titleFontSize)
       .attr("x", (d,i) => { return xBuffer - cubeWidth*3 - i*cubeWidth*5 })
       .attr("y", yBuffer)
       .text((d) => { return d });

  items.data(numbers)
       .enter()
       .append("rect")
       .attr("x", (d,i) => { return xBuffer - cubeWidth*3 - i*cubeWidth*5 })
       .attr("y", yBuffer+0.4*cubeWidth)
       .attr("width", cubeWidth * 4)
       .attr("height", cubeWidth);

  items.data(numbers)
       .enter()
       .append("text")
       .style("font-size", elemFontSize)
       .text((d) => {
         var result = d;
         if (typeof(d) === "number") {
          result = d.toString(16); // mostly this will be bits, but if not hex it
         }
         return result;
       })
       .attr("x", (d,i) => { return xBuffer - cubeWidth*3 - i*cubeWidth*5 + cubeWidth/4 })
       .attr("y", yBuffer + 1.10*cubeWidth);
}

let appendLegend = function(titles, colours, yIncrement){
  var items = svgDoc.selectAll("g");

  items.data(colours)
       .enter()
       .append("rect")
       .style("fill", (d) => { return d })
       .attr("x", (d,i) => { return (xBuffer - i*cubeWidth*5 - cubeWidth*3 ) })
       .attr("y", yBuffer)
       .attr("width", 4*cubeWidth)
       .attr("height", cubeWidth);

  items.data(titles)
       .enter()
       .append("text")
       .style("font-size", elemFontSize)
       .text((d) => { return d })
       .attr("x", (d,i) => { return (xBuffer - i*cubeWidth*5 - cubeWidth*2) })
       .attr("y", yBuffer + 0.60*cubeWidth);

  yBuffer+=yIncrement*cubeWidth;
}

let noop = function(d, i) { return null };

let input = function(hits, doubleHits, pos) {
    return function(d, i) {
        let ctr=0;

        if (hits.includes(i)) {
          ctr+=1;
        }

        if (doubleHits.includes(i)) {
          ctr+=2;
        }

        if (ctr > 0) {
          return hitColours[ctr-1];
        }

        if (i == pos) {
          return counterColour;
        }
    };
};

let newHash = function(done){
  inputText = $("#algo-input")[0].value;
  inputBytes = strToByteArr(inputText);

  $.ajax ({
        url: "/ctph/init",
        type: "POST",
        data: JSON.stringify({"data_length":inputBytes.length}),
        dataType: "json",
        contentType: "application/json; charset=utf-8"
  }).fail(function(a,b,c){
    console.log(a,b,c);
    console.log("failed");
  }).done(function(data){
    done(JSON.parse(data))
  });
}

let render = function(){
  let dBits = bitArray([inputBytes[ctr]]);
  yBuffer=1*cubeWidth;

  svgDoc.html(null);

  appendLegend(["ModBS", "Mod2BS", "Both"], hitColours,3);
  appendArray("Input Text", inputText, input(hits, doubleHits, ctr),3);
  appendArray("Input Bytes (hex)", inputBytes, input(hits, doubleHits, ctr),3);
  appendArray("Bits of current selection (d)", dBits.slice(0,8),noop, 3);
  appendArray("Window Array (hex)", fh.rolling_hash.window, noop, 3);

  appendText(["Z Value", "Y Value", "X Value"],
    [fh.rolling_hash.z.toString(10),
      fh.rolling_hash.y.toString(10),
      fh.rolling_hash.x.toString(10)]);

  let sig = fh.block_size + ":" + fh.sig1 + ":" + fh.sig2;
  document.getElementById("algo-output").value = sig;
};

let stepHash = function(){
  ctr++;

  // Block while we wait for server response
  // Obviosuly this is a cosmetic block, but this isn't some security check
  // If someone wants to gun via the console, then they just get wrong ssdeep results :P
  document.getElementById("button2").disabled = true;

  $.ajax ({
        url: "/ctph/step",
        type: "POST",
        data: JSON.stringify({"byte":inputBytes[ctr]}),
        dataType: "json",
        contentType: "application/json; charset=utf-8"
  }).fail(function(a,b,c){
    console.log(a,b,c);
    console.log("failed");
  }).done(function(data){
    data = JSON.parse(data);

    if (null != data) {
      fh = data // GLOBAL
      if (fh.is_trigger1) {
        hits.push(ctr);
      }
      if (fh.is_trigger2) {
        doubleHits.push(ctr);
      }
      render();
    }
  }).always(function(){
    // Renable the button
    document.getElementById("button2").disabled = false;
  });
}

let init = function() {

  // GLOBALS
  newHash((data) => {
    fh = data;
    ctr = fh.index;
    hits = [];
    doubleHits = [];

    updateSizing();
    render();
  });
}

window.onresize = function(event) {
  updateSizing();
  render();
};

init();
