Example usage of the Arbiter package.
along with an example of an alternative naive mutex-based strategies (for comparitive benchmarking).

Example use case:
* Idempotent application of a configuration set to an environment, which is tracked by a unique, always increasing version number (eg version '2' supersedes and replaces version '1', etc).  This version is tracked in a keystore to improve performance (no need to query 'on disk' configuration to obtain currently installed version #).  Updates contain all information needed, so applying version '5' over version '3' does not incur any loss of configuration data (thus configuration application is idempotent).
* If the application of the configuration set to the environment fails, the version being stored in the tracking keystore is not updated, and error returned to caller.  For a deterministic simulation of intermittent failures, if the version number is prime, it fails to update.
* If an older or redundant (equal to current) version update comes in (due to delayed/ duplicate processing upstream), the config update is rejected, and an error returned to the caller.

Benchmarking Results:
```
goos: linux
goarch: amd64
pkg: github.com/btsomogyi/arbiter/example/arbitrated
cpu: Intel(R) Core(TM) i7-10750H CPU @ 2.60GHz
Benchmark_randomRequests/100_reqs_@_1-12         	      30	  41670903 ns/op	 3445678 B/op	   23352 allocs/op
Benchmark_randomRequests/100_reqs_@_2-12         	      51	  25299301 ns/op	 3343849 B/op	   22887 allocs/op
Benchmark_randomRequests/100_reqs_@_4-12         	      79	  15644064 ns/op	 3445646 B/op	   22518 allocs/op
Benchmark_randomRequests/100_reqs_@_8-12         	     100	  10019367 ns/op	 3242817 B/op	   22175 allocs/op
Benchmark_randomRequests/100_reqs_@_12-12        	     132	   9143837 ns/op	 3390791 B/op	   22041 allocs/op
Benchmark_randomRequests/100_reqs_@_24-12        	     150	   8551522 ns/op	 3446120 B/op	   21974 allocs/op
Benchmark_randomRequests/500_reqs_@_1-12         	       5	 207684064 ns/op	 8523768 B/op	  115799 allocs/op
Benchmark_randomRequests/500_reqs_@_2-12         	       9	 122748809 ns/op	 8335147 B/op	  113267 allocs/op
Benchmark_randomRequests/500_reqs_@_4-12         	      15	  73492129 ns/op	 8507206 B/op	  111195 allocs/op
Benchmark_randomRequests/500_reqs_@_8-12         	      26	  46641672 ns/op	 8573232 B/op	  109316 allocs/op
Benchmark_randomRequests/500_reqs_@_12-12        	      31	  39829528 ns/op	 8605288 B/op	  108473 allocs/op
Benchmark_randomRequests/500_reqs_@_24-12        	      33	  34827456 ns/op	 8592491 B/op	  107644 allocs/op
Benchmark_randomRequests/500_reqs_@_48-12        	      31	  33366654 ns/op	 8597645 B/op	  107519 allocs/op
Benchmark_randomRequests/5000_reqs_@_1-12        	       1	2127903352 ns/op	64822240 B/op	 1159295 allocs/op
Benchmark_randomRequests/5000_reqs_@_2-12        	       1	1276234308 ns/op	65330256 B/op	 1130826 allocs/op
Benchmark_randomRequests/5000_reqs_@_4-12        	       2	 750219968 ns/op	64962672 B/op	 1110850 allocs/op
Benchmark_randomRequests/5000_reqs_@_8-12        	       3	 451222826 ns/op	64592517 B/op	 1091353 allocs/op
Benchmark_randomRequests/5000_reqs_@_12-12       	       3	 371467099 ns/op	64455080 B/op	 1081698 allocs/op
Benchmark_randomRequests/5000_reqs_@_24-12       	       4	 324574178 ns/op	64310772 B/op	 1073173 allocs/op
Benchmark_randomRequests/5000_reqs_@_48-12       	       4	 305642515 ns/op	64252070 B/op	 1069895 allocs/op
Benchmark_randomRequests/5000_reqs_@_96-12       	       4	 302397088 ns/op	64244590 B/op	 1068821 allocs/op
Benchmark_randomRequests/5000_reqs_@_144-12      	       4	 297982866 ns/op	64214630 B/op	 1068221 allocs/op
Benchmark_randomRequests/5000_reqs_@_192-12      	       4	 296059368 ns/op	64247506 B/op	 1068318 allocs/op
Benchmark_randomRequests/50000_reqs_@_24-12         	       1	3035914957 ns/op	619352608 B/op	10737529 allocs/op
Benchmark_randomRequests/50000_reqs_@_48-12         	       1	2861134912 ns/op	618344984 B/op	10705097 allocs/op
Benchmark_randomRequests/50000_reqs_@_96-12         	       1	2797182453 ns/op	618037576 B/op	10691244 allocs/op
Benchmark_randomRequests/50000_reqs_@_144-12        	       1	2799502403 ns/op	617689320 B/op	10685973 allocs/op
Benchmark_randomRequests/50000_reqs_@_192-12        	       1	2784231075 ns/op	617891016 B/op	10685157 allocs/op
Benchmark_randomRequests/50000_reqs_@_384-12        	       1	2789885225 ns/op	618106768 B/op	10683443 allocs/op
Benchmark_randomRequests/50000_reqs_@_768-12        	       1	2773276816 ns/op	618361736 B/op	10683146 allocs/op
PASS
ok  	github.com/btsomogyi/arbiter/example/arbitrated	43.012s
```
```
goos: linux
goarch: amd64
pkg: github.com/btsomogyi/arbiter/example/locking
cpu: Intel(R) Core(TM) i7-10750H CPU @ 2.60GHz
Benchmark_randomRequests/100_reqs_@_1-12         	      37	  28610604 ns/op	 3116350 B/op	   18186 allocs/op
Benchmark_randomRequests/100_reqs_@_2-12         	      64	  17652793 ns/op	 2931866 B/op	   17518 allocs/op
Benchmark_randomRequests/100_reqs_@_4-12         	      96	  10497181 ns/op	 2997057 B/op	   17148 allocs/op
Benchmark_randomRequests/100_reqs_@_8-12         	     150	   8149541 ns/op	 3018817 B/op	   16889 allocs/op
Benchmark_randomRequests/100_reqs_@_12-12        	     153	   7906349 ns/op	 3142623 B/op	   16844 allocs/op
Benchmark_randomRequests/100_reqs_@_24-12        	     154	   7800273 ns/op	 3125709 B/op	   16838 allocs/op
Benchmark_randomRequests/500_reqs_@_1-12         	       7	 145186026 ns/op	 7068329 B/op	   89844 allocs/op
Benchmark_randomRequests/500_reqs_@_2-12         	      13	  85512014 ns/op	 6841974 B/op	   86005 allocs/op
Benchmark_randomRequests/500_reqs_@_4-12         	      24	  49466714 ns/op	 6833347 B/op	   83900 allocs/op
Benchmark_randomRequests/500_reqs_@_8-12         	      33	  36896728 ns/op	 6810256 B/op	   82507 allocs/op
Benchmark_randomRequests/500_reqs_@_12-12        	      36	  34802250 ns/op	 6906511 B/op	   82071 allocs/op
Benchmark_randomRequests/500_reqs_@_24-12        	      36	  32070577 ns/op	 6843102 B/op	   81651 allocs/op
Benchmark_randomRequests/500_reqs_@_48-12        	      34	  31210601 ns/op	 6932942 B/op	   81571 allocs/op
Benchmark_randomRequests/5000_reqs_@_1-12        	       1	1477490890 ns/op	49078184 B/op	  896370 allocs/op
Benchmark_randomRequests/5000_reqs_@_2-12        	       2	 841370130 ns/op	48274288 B/op	  853604 allocs/op
Benchmark_randomRequests/5000_reqs_@_4-12        	       3	 475446437 ns/op	47864088 B/op	  832406 allocs/op
Benchmark_randomRequests/5000_reqs_@_8-12        	       3	 355234679 ns/op	47597786 B/op	  818040 allocs/op
Benchmark_randomRequests/5000_reqs_@_12-12       	       4	 328360500 ns/op	47553708 B/op	  813177 allocs/op
Benchmark_randomRequests/5000_reqs_@_24-12       	       4	 300473129 ns/op	47466486 B/op	  808396 allocs/op
Benchmark_randomRequests/5000_reqs_@_48-12       	       4	 291820042 ns/op	47423374 B/op	  806304 allocs/op
Benchmark_randomRequests/5000_reqs_@_96-12       	       4	 287728296 ns/op	47412988 B/op	  805262 allocs/op
Benchmark_randomRequests/5000_reqs_@_144-12      	       4	 285370240 ns/op	47382480 B/op	  804703 allocs/op
Benchmark_randomRequests/5000_reqs_@_192-12      	       4	 284993645 ns/op	47411646 B/op	  804887 allocs/op
Benchmark_randomRequests/50000_reqs_@_24-12         	       1	2877858778 ns/op	450228480 B/op	 8058232 allocs/op
Benchmark_randomRequests/50000_reqs_@_48-12         	       1	2805582166 ns/op	449451064 B/op	 8034379 allocs/op
Benchmark_randomRequests/50000_reqs_@_96-12         	       1	2726901612 ns/op	449132728 B/op	 8022565 allocs/op
Benchmark_randomRequests/50000_reqs_@_144-12        	       1	2716296790 ns/op	448767488 B/op	 8017602 allocs/op
Benchmark_randomRequests/50000_reqs_@_192-12        	       1	2717129807 ns/op	448958320 B/op	 8016662 allocs/op
Benchmark_randomRequests/50000_reqs_@_384-12        	       1	2687523806 ns/op	449048672 B/op	 8014978 allocs/op
Benchmark_randomRequests/50000_reqs_@_768-12        	       1	2686173000 ns/op	449203408 B/op	 8015036 allocs/op
PASS
ok  	github.com/btsomogyi/arbiter/example/locking	42.011s
```

```
| |Arbitrated|	|	|	|	|Locking|	|	|	|	|Ratio Arbitrated/Locking	|	|	|
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
|benchmark|requests|concurrency|operations|ns/op|B/op|allocs/op|operations|ns/op|B/op|allocs/op|operations|ns/op|B/op|allocs/op|
|100_reqs_@_1-12|100|1|30|41670903|3445678|23352|37|28610604|3116350|18186|0.8108108108|1.456484561|1.105677475|1.284064665|
|100_reqs_@_2-12|100|2|51|25299301|3343849|22887|64|17652793|2931866|17518|0.796875|1.43316137|1.140519041|1.306484759|
|100_reqs_@_4-12|100|4|79|15644064|3445646|22518|96|10497181|2997057|17148|0.8229166667|1.49031097|1.149676499|1.313156053|
|100_reqs_@_8-12|100|8|100|10019367|3242817|22175|150|8149541|3018817|16889|0.6666666667|1.229439425|1.074201252|1.312984783|
|100_reqs_@_12-12|100|12|132|9143837|3390791|22041|153|7906349|3142623|16844|0.862745098|1.156518261|1.078968429|1.308537165|
|100_reqs_@_24-12|100|24|150|8551522|3446120|21974|154|7800273|3125709|16838|0.974025974|1.096310603|1.102508263|1.30502435|
|500_reqs_@_1-12|500|1|5|207684064|8523768|115799|7|145186026|7068329|89844|0.7142857143|1.430468687|1.205909912|1.288889631|
|500_reqs_@_2-12|500|2|9|122748809|8335147|113267|13|85512014|6841974|86005|0.6923076923|1.435456882|1.218237164|1.316981571|
|500_reqs_@_4-12|500|4|15|73492129|8507206|111195|24|49466714|6833347|83900|0.625|1.485688518|1.244954486|1.325327771|
|500_reqs_@_8-12|500|8|26|46641672|8573232|109316|33|36896728|6810256|82507|0.7878787879|1.264114043|1.258870738|1.324930006|
|500_reqs_@_12-12|500|12|31|39829528|8605288|108473|36|34802250|6906511|82071|0.8611111111|1.144452672|1.245967465|1.321697067|
|500_reqs_@_24-12|500|24|33|34827456|8592491|107644|36|32070577|6843102|81651|0.9166666667|1.085962875|1.25564269|1.318342702|
|500_reqs_@_48-12|500|48|31|33366654|8597645|107519|34|31210601|6932942|81571|0.9117647059|1.069080791|1.240114947|1.318103247|
|5000_reqs_@_1-12|5000|1|1|2127903352|64822240|1159295|1|1477490890|49078184|896370|1|1.440214194|1.320795407|1.293321954|
|5000_reqs_@_2-12|5000|2|1|1276234308|65330256|1130826|2|841370130|48274288|853604|0.5|1.516852408|1.353313714|1.324766519|
|5000_reqs_@_4-12|5000|4|2|750219968|64962672|1110850|3|475446437|47864088|832406|0.6666666667|1.577927416|1.357232002|1.334505037|
|5000_reqs_@_8-12|5000|8|3|451222826|64592517|1091353|3|355234679|47597786|818040|1|1.270210519|1.357048771|1.334107134|
|5000_reqs_@_12-12|5000|12|3|371467099|64455080|1081698|4|328360500|47553708|813177|0.75|1.131278272|1.35541649|1.330212242|
|5000_reqs_@_24-12|5000|24|4|324574178|64310772|1073173|4|300473129|47466486|808396|1|1.080210331|1.354866927|1.327533783|
|5000_reqs_@_48-12|5000|48|4|305642515|64252070|1069895|4|291820042|47423374|806304|1|1.047366428|1.354860791|1.326912678|
|5000_reqs_@_96-12|5000|96|4|302397088|64244590|1068821|4|287728296|47412988|805262|1|1.050981402|1.354999816|1.327295961|
|5000_reqs_@_144-12|5000|144|4|297982866|64214630|1068221|4|285370240|47382480|804703|1|1.044197412|1.355239954|1.327472372|
|5000_reqs_@_192-12|5000|192|4|296059368|64247506|1068318|4|284993645|47411646|804887|1|1.038827964|1.355099673|1.327289421|
|50000_reqs_@_24-12|50000|24|1|3035914957|619352608|10737529|1|2877858778|450228480|8058232|1|1.054921451|1.37564067|1.332491916|
|50000_reqs_@_48-12|50000|48|1|2861134912|618344984|10705097|1|2805582166|449451064|8034379|1|1.019800791|1.375778218|1.332411254|
|50000_reqs_@_96-12|50000|96|1|2797182453|618037576|10691244|1|2726901612|449132728|8022565|1|1.025773149|1.376068893|1.332646604|
|50000_reqs_@_144-12|50000|144|1|2799502403|617689320|10685973|1|2716296790|448767488|8017602|1|1.030632004|1.376412812|1.332814101|
|50000_reqs_@_192-12|50000|192|1|2784231075|617891016|10685157|1|2717129807|448958320|8016662|1|1.024695643|1.376277014|1.332868593|
|50000_reqs_@_384-12|50000|384|1|2789885225|618106768|10683443|1|2687523806|449048672|8014978|1|1.038087632|1.376480561|1.332934788|
|50000_reqs_@_768-12|50000|768|1|2773276816|618361736|10683146|1|2686173000|449203408|8015036|1|1.032426733|1.376574009|1.332888087|
```
