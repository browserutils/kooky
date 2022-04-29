# Notes

## Version History

Internet Explorer 4 uses index.dat URL cache v4.7
Internet Explorer 9 uses index.dat
Internet Explorer 10, 11 use WebCacheV01.dat

https://www.digital-detective.net/random-cookie-filenames/
"To mitigate the threat, Internet Explorer 9.0.2 now names the cookie files using a randomly-generated alphanumeric string."

http://hh.diva-portal.org/smash/get/diva2:635743/FULLTEXT02.pdf p6
"With Internet Explorer 10, Microsoft changed the way of storing web related information.
Instead of the old index.dat files, Internet Explorer 10 uses an ESE database called WebCacheV01.dat" ...

https://www.nirsoft.net/utils/edge_cookies_view.html
starting from Fall Creators Update 1709 of Windows 10, the cookies of Microsoft Edge Web browser are stored in the WebCacheV01.dat database
ESE database at %USERPROFILE%\AppData\Local\Microsoft\Windows\WebCache\WebCacheV01.dat (%LocalAppData%\Microsoft\Windows\WebCache\WebCacheV01.dat)
CookieEntryEx_##

https://www.foxtonforensics.com/browser-history-examiner/microsoft-edge-history-location
v79+:
Edge Cookies are stored in the 'Cookies' SQLite database, within the 'cookies' table.

up to Edge v44:
Edge Cookies are stored in the 'WebCacheV01.dat' ESE database, within the 'CookieEntryEx' containers.

TODO: Older versions of Edge stored cookies as separate text files in locations specified within the ESE database. (?)

## Text Cookies

### Format

http://index-of.es/Forensic/Forensic%20Analysis%20of%20Microsoft%20Internet%20Explorer%20Cookie%20Files.pdf # least and most significant switched

https://master.dl.sourceforge.net/project/odessa/ODESSA/White%20Papers/IE_Cookie_File_Reconstruction.pdf?viasf=1

https://www.consumingexperience.com/2011/09/internet-explorer-cookie-contents-new.html

### Code

https://www.codeproject.com/Articles/330142/Cookie-Quest-A-Quest-to-Read-Cookies-from-Four-Pop#InternetExplorer1

https://sourceforge.net/projects/odessa/files/Galleta/20040505_1/galleta_20040505_1.tar.gz/download # wrong times

## IE URL Cache Cookies

### Format

https://github.com/libyal/libmsiecf/blob/main/documentation/MSIE%20Cache%20File%20(index.dat)%20format.asciidoc

https://www.geoffchappell.com/studies/windows/ie/wininet/api/urlcache/indexdat.htm

https://web.archive.org/web/20170427044924/http://www.latenighthacking.com/projects/2003/reIndexDat/

https://web.archive.org/web/20020811061516/http://www.conknet.com/~w_kranz/mswinbrz.txt

https://master.dl.sourceforge.net/project/odessa/OldFiles/Analysis_IEHistory.pdf?viasf=1

http://www.stevebunting.org/udpd4n6/forensics/index_dat1.htm

https://tzworks.com/prototypes/index_dat/id.users.guide.pdf

## Code

https://github.com/bauman/python-pasco/blob/master/pasco/pascohelpermodule.c

https://www.geoffchappell.com/studies/windows/ie/wininet/api/urlcache/hashkey.htm?tx=20,78,83,84,88

https://doxygen.reactos.org/da/df8/dll_2win32_2wininet_2urlcache_8c_source.html # ReactOS

### Tools

https://web.archive.org/web/20201006212958/https://tzworks.net/prototype_page.php?proto_id=6

https://www.nirsoft.net/utils/iecookies.html

## ESE Database Cookies

### Format

http://hh.diva-portal.org/smash/get/diva2:635743/FULLTEXT02.pdf

https://raw.githubusercontent.com/libyal/documentation/main/Forensic%20analysis%20of%20the%20Windows%20Search%20database.pdf

https://github.com/libyal/libesedb/blob/main/documentation/Extensible%20Storage%20Engine%20(ESE)%20Database%20File%20(EDB)%20format.asciidoc

https://web.archive.org/web/20191220154919/https://bsmuir.kinja.com/windows-10-microsoft-edge-browser-forensics-1733533818

https://www.linkedin.com/pulse/windows-10-microsoft-edge-browser-forensics-brent-muir


### Tools

https://www.nirsoft.net/utils/edge_cookies_view.html

https://www.nirsoft.net/utils/ese_database_view.html

https://www.digital-detective.net/dcode/ # time stamps

## Test Environments

https://web.archive.org/web/20150410044551/https://www.modern.ie/en-us # test VMs

https://developer.microsoft.com/en-us/microsoft-edge/tools/vms/ # test VMs
