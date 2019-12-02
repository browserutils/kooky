# A Go based ESE parser.

The Extensible Storage Engine (ESE) Database File is commonly used
within Windows to store various application specific information. It
is the Microsoft analogue to sqlite - so just like sqlite is used to
store chrome history, ESE is used to store Internet Explorer history.

In essence it is a flat file database. This project is a library to
help read such a file. The following description is a high lieve
account of the main feautures of the file format and how to access
these using the library.

## File format overview

The file consists of pages. The page size can vary but it is specified
in the file header.

The file may contain multiple objects (tables) stored within
pages. The pages form a B tree where data is stored in the actual leaf
pages. As the file grows the tree canbe extended by inserting pages
into it.

Data is stored inside each page in a `Tag`. Tags are just a series of
data blobs stored in each page.

The WalkPages() function can be used to produce all the tags starting
at a root page. The function walks the B+ tree automatically and
parses out each tag. The callback is called for each tag - if the
callback returns an error the walk is stopped and the error is relayed
to the WalkPages() caller.

## The page

Each page contains a series of tags. You can see help about a specific
page using the `page` command:

```
$ eseparser page WebCacheV01.dat  4
Page 4: struct PageHeader @ 0x28000:
  LastModified: {
  struct DBTime @ 0x28008:
    Hours: 0xfbf4
    Min: 0x0
    Sec: 0x0
  }
  PreviousPageNumber: 0x0
  NextPageNumber: 0x0
  FatherPage: 0x2
  AvailableDataSize: 0x7f3b
  AvailableDataOffset: 0x5d
  AvailablePageTag: 0x6
  Flags: 108549 (Parent,Root)

Tag 0 @ 0x2fffc offset 0x0 length 0x10
Tag 1 @ 0x2fff8 offset 0x16 length 0x13
Tag 2 @ 0x2fff4 offset 0x29 length 0xe
Tag 3 @ 0x2fff0 offset 0x37 length 0x13
Tag 4 @ 0x2ffec offset 0x4a length 0x13
Tag 5 @ 0x2ffe8 offset 0x10 length 0x6
struct ESENT_ROOT_HEADER @ 0x0:
  InitialNumberOfPages: 0x14
  ParentFDP: 0x1
  ExtentSpace: Multiple (1)
  SpaceTreePageNumber: 0x5

struct ESENT_BRANCH_ENTRY @ 0x0:
  LocalPageKeySize: 0xd
  ChildPageNumber: 0xd
struct ESENT_BRANCH_ENTRY @ 0x0:
  LocalPageKeySize: 0x8
  ChildPageNumber: 0xe
struct ESENT_BRANCH_ENTRY @ 0x0:
  LocalPageKeySize: 0xd
  ChildPageNumber: 0x13
struct ESENT_BRANCH_ENTRY @ 0x0:
  LocalPageKeySize: 0xd
  ChildPageNumber: 0x14
struct ESENT_BRANCH_ENTRY @ 0x0:
  LocalPageKeySize: 0x0
  ChildPageNumber: 0x16
```

The example above shows a root page (4) containing 5 branch nodes.

## The catalog

The ESE file contains a catalog starting from page 4. The catalog
defines all the tables, their columns and types stat are stored in the
database.

You can see the catalong by runing the `catalog` command:

```
$ eseparser catalog /shared/WebCacheV01.dat
[MSysObjects] (FDP 0x4):
   Columns
      0    ObjidTable                    Signed long
      1    Type                          Signed short
      2    Id                            Signed long
      3    ColtypOrPgnoFDP               Signed long
      4    SpaceUsage                    Signed long
      5    Flags                         Signed long
      6    PagesOrLocale                 Signed long
      7    RootFlag                      Boolean
      8    RecordOffset                  Signed short
```

The first table in the catalog called `MSysObjects` is really a
database table containing a description of all the tables in the file.

## Tables

Ultimately the ESE format is a database storage engine and it stores
rows in tables. Each table is stored inside the B tree rooted by the
DFP ID shown in the catalog.  Each row is stored inside a tag (inside
one of the pages within the tree).

There are three types of columns:

- Fixed size (e.g. integers) have a known size.
- Variable size (e.g. Strings) have a variable size.
- Tagged data - these columns are often null and therefore may not be
  present. The database stored these with their column ID as a map.

Therefore within the tag for each column, there are three distinct
storage areas. You can see how each record is parsed using the --debug flag:

```
$ eseparser dump WebCacheV01.dat MSysObjects --debug
Walking page 4
Got 6 values for page 4
Walking page 13
Got 404 values for page 13
Processing row in Tag @ 491512 0xd (0x37)([]uint8) (len=55 cap=55) {
 00000000  07 00 06 00 01 7f 80 00  00 02 08 80 20 00 02 00  |............ ...|
 00000010  00 00 01 00 02 00 00 00  04 00 00 00 50 00 00 00  |............P...|
 00000020  00 00 00 c0 14 00 00 00  ff 00 0b 00 4d 53 79 73  |............MSys|
 00000030  4f 62 6a 65 63 74 73                              |Objects|
}
struct ESENT_LEAF_ENTRY @ 0x2:
  CommonPageKeySize: 0x7
  LocalPageKeySize: 0x6

struct ESENT_DATA_DEFINITION_HEADER @ 0xa:
  LastFixedType: 0x8
  LastVariableDataType: 0x80
  VariableSizeOffset: 0x20

Column ObjidTable Identifier 1 Type Signed long
Consume 0x4 bytes of FIXED space from 0xe
Column Type Identifier 2 Type Signed short
Consume 0x2 bytes of FIXED space from 0x12
Column Id Identifier 3 Type Signed long
...
{"ObjidTable":2,"Type":1,"Id":2,"ColtypOrPgnoFDP":4,"SpaceUsage":80,"Flags":-1073741824,"PagesOrLocale":20,"RootFlag":true,"Name":"MSysObjects"}
```

If you just want to dump out the columns omit the `--debug` flag:
```
$ ./eseparser dump /shared/WebCacheV01.dat HstsEntryEx_46
{"EntryId":1,"MinimizedRDomainHash":0,"MinimizedRDomainLength":8,"IncludeSubdomains":177,"Expires":9223372036854775807,"LastTimeUsed":9223372036854775807,"RDomain":":version"}
{"EntryId":2,"MinimizedRDomainHash":1536723792475384667,"MinimizedRDomainLength":10,"IncludeSubdomains":1,"Expires":132508317752329580,"LastTimeUsed":132192957752329580,"RDomain":"com.office.www"}
```



References:
 * https://github.com/SecureAuthCorp/impacket.git
 * https://blogs.technet.microsoft.com/askds/2017/12/04/ese-deep-dive-part-1-the-anatomy-of-an-ese-database/
 * https://docs.microsoft.com/en-us/previous-versions/windows/it-pro/windows-server-2012-R2-and-2012/hh875546(v=ws.11)?redirectedfrom=MSDN
 * http://hh.diva-portal.org/smash/get/diva2:635743/FULLTEXT02.pdf
 * https://github.com/libyal/libesedb/blob/master/documentation/Extensible%20Storage%20Engine%20(ESE)%20Database%20File%20(EDB)%20format.asciidoc
