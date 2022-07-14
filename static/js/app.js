const algoInputElement = $("#algo-input")[0];
const algoOutputElement = $("#algo-output").get(0);
let counterColour;
let ctr = 0;
if (counterColour === undefined) {
    counterColour = "chartreuse";
    const hitColours = ["red", "dodgerblue", "indigo"];
    const elemFontSize = "0.22em", titleFontSize = "0.22em";
    var cubeWidth = 10, xBuffer = 290, yBuffer = 0, svgDoc = d3.selectAll("svg");
    var updateSizing = () => {
        let inputBytes = strToByteArr(algoInputElement.value);
        if (typeof inputBytes !== 'undefined') {
            cubeWidth = xBuffer / inputBytes.length;
        }
    };
    function strToByteArr(str) {
        var arr = [];
        for (var i = 0; i < str.length; i++) {
            arr.push(str.charCodeAt(i));
        }
        return arr;
    }
    function bitArray(arr) {
        var output = []; // there is no bit array :(
        for (var i = 0; i < arr.length; i++) {
            for (var j = 0; j < 32; j++) {
                var mask = 1 << j;
                if ((arr[i] & mask) == mask) {
                    output.push(1);
                }
                else {
                    output.push(0);
                }
            }
        }
        return output;
    }
    function appendArray(title, backingArray, highlight) {
        var items = svgDoc.selectAll("g");
        items.data([title])
            .enter()
            .append("text")
            .style("font-size", titleFontSize)
            .attr("x", xBuffer + cubeWidth)
            .style("text-anchor", "end")
            .attr("y", yBuffer)
            .text(d => d);
        items.data(backingArray)
            .enter()
            .append("rect")
            .attr("x", (d, i) => { return (xBuffer - i * cubeWidth); })
            .attr("y", yBuffer + 0.4 * cubeWidth)
            .attr("width", cubeWidth)
            .attr("height", cubeWidth)
            .style("fill", highlight);
        items.data(backingArray)
            .enter()
            .append("text")
            .text((d) => {
            var result = d;
            if (typeof (d) === "number") {
                result = d.toString(16); // mostly this will be bits, but if not hex it
            }
            return result;
        })
            .style("font-size", elemFontSize)
            .attr("x", (d, i) => { return (xBuffer - (i - 0.5) * cubeWidth); })
            .attr("text-anchor", "middle")
            .attr("y", yBuffer + cubeWidth)
            .attr("dominant-baseline", "middle");
    }
    function appendText(titles, numbers) {
        var items = svgDoc.selectAll("g");
        items.data(titles)
            .enter()
            .append("text")
            .style("font-size", titleFontSize)
            .attr("x", (d, i) => { return xBuffer - cubeWidth * 3 - i * cubeWidth * 5; })
            .attr("y", yBuffer)
            .text((d) => { return d; });
        items.data(numbers)
            .enter()
            .append("rect")
            .attr("x", (d, i) => { return xBuffer - cubeWidth * 3 - i * cubeWidth * 5; })
            .attr("y", yBuffer + 0.4 * cubeWidth)
            .attr("width", cubeWidth * 4)
            .attr("height", cubeWidth);
        items.data(numbers)
            .enter()
            .append("text")
            .style("font-size", elemFontSize)
            .text((d) => {
            var result = d;
            if (typeof (d) === "number") {
                result = d.toString(16); // mostly this will be bits, but if not hex it
            }
            return result;
        })
            .attr("x", (d, i) => { return xBuffer - cubeWidth * 3 - i * cubeWidth * 5 + cubeWidth / 4; })
            .attr("y", yBuffer + 1.10 * cubeWidth)
            .attr("dominant-baseline", "middle");
    }
    function appendLegend(titles, colours) {
        let items = svgDoc.selectAll("g");
        items.data(colours)
            .enter()
            .append("rect")
            .style("fill", (d) => { return d; })
            .attr("x", (d, i) => { return (xBuffer - i * cubeWidth * 5 - cubeWidth * 3); })
            .attr("y", yBuffer)
            .attr("width", 4 * cubeWidth)
            .attr("height", cubeWidth);
        items.data(titles)
            .enter()
            .append("text")
            .style("font-size", elemFontSize)
            .attr("dominant-baseline", "middle")
            .text((d) => { return d; })
            .attr("x", (d, i) => { return (xBuffer - i * cubeWidth * 5 - cubeWidth * 2); })
            .attr("y", yBuffer + 0.60 * cubeWidth);
    }
    function noop(d, i) {
        return null;
    }
    function input(hits, doubleHits, pos) {
        return function (d, i) {
            ctr = 0;
            if (hits.includes(i)) {
                ctr += 1;
            }
            if (doubleHits.includes(i)) {
                ctr += 2;
            }
            if (ctr > 0) {
                return hitColours[ctr - 1];
            }
            if (i == pos) {
                return counterColour;
            }
        };
    }
    function render() {
        let inputText = algoInputElement.value;
        let inputBytes = strToByteArr(inputText);
        var dBits = bitArray([inputBytes[ctr]]);
        yBuffer = 0;
        svgDoc.html(null);
        appendLegend(["ModBS", "Mod2BS", "Both"], hitColours);
        yBuffer += 2 * cubeWidth;
        appendArray("Input Text", inputText, input(hits, doubleHits, ctr));
        yBuffer += 3 * cubeWidth;
        appendArray("Input Bytes (hex)", inputBytes, input(hits, doubleHits, ctr));
        yBuffer += 3 * cubeWidth;
        appendArray("Bits of current selection (d)", dBits.slice(0, 8), noop);
        yBuffer += 3 * cubeWidth;
        appendArray("Window Array (hex)", fh.rolling_hash.window, noop);
        yBuffer += 3 * cubeWidth;
        appendText(["Z Value", "Y Value", "X Value"], [fh.rolling_hash.z.toString(10),
            fh.rolling_hash.y.toString(10),
            fh.rolling_hash.x.toString(10)]);
        var sig = fh.block_size + ":" + fh.sig1 + ":" + fh.sig2;
        algoOutputElement.value = sig;
    }
    ;
    function stepAlgo() {
        var algoPath = localStorage.getItem("algoPathName");
        if (algoPath == null) {
            return;
        }
        var inputText = algoInputElement.value;
        var inputBytes = strToByteArr(inputText);
        ctr++;
        $.ajax({
            async: false,
            contentType: "application/json; charset=utf-8",
            data: JSON.stringify({ byte: inputBytes[ctr] }),
            dataType: "json",
            type: "POST",
            url: `${algoPath}/step`,
        })
            .fail(function (a, b, c) {
            console.log(a, b, c);
            console.log("failed");
        })
            .done(function (data) {
            data = JSON.parse(data);
            if (null != data) {
                fh = data; // GLOBAL
                if (fh.is_trigger1) {
                    hits.push(ctr);
                }
                if (fh.is_trigger2) {
                    doubleHits.push(ctr);
                }
                render();
            }
        });
    }
    function initAlgo() {
        var algoPath = localStorage.getItem("algoPathName");
        if (algoPath == null) {
            return;
        }
        // stick with var throughout for these two, over var
        // in case of redefine being an issue
        var inputText = algoInputElement.value;
        var inputBytes = strToByteArr(inputText);
        $.ajax({
            async: false,
            contentType: "application/json; charset=utf-8",
            data: JSON.stringify({ data_length: inputBytes.length }),
            dataType: "json",
            type: "POST",
            url: `${algoPath}/init`,
        })
            .fail(function (a, b, c) {
            console.log(a, b, c);
            console.log("failed");
        })
            .done(function (response) {
            var parsed = JSON.parse(response);
            // GLOBALS
            fh = parsed;
            ctr = fh.index;
            hits = [];
            doubleHits = [];
            updateSizing();
            render();
        });
    }
    window.onresize = () => {
        updateSizing();
        render();
    };
    $(document).ready(() => {
        initAlgo();
    });
}
export {};
