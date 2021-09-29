package main

import "log"

// SQL to generate this:
//
// SELECT ARRAY_TO_STRING(ARRAY_AGG(DISTINCT CAST(FieldID AS STRING)), ",\n", "") fieldids
// FROM `ukbb-analyses.ukbb7089_201910.materialized_special_dates`

func init() {
	knownSpecialFields := []int{
		20002,
		20004,
		40006,
		6152,
		131350,
		131891,
		130070,
		131308,
		131037,
		20001,
		40021,
		40013,
		131682,
		42007,
		42001,
		6150,
		131632,
		131369,
		131028,
		131324,
		131396,
		131307,
		131423,
		131457,
		130071,
		40020,
		131347,
		131306,
		131456,
		131309,
		131288,
		131692,
		132050,
		131310,
		42027,
		131289,
		130651,
		131658,
		131368,
		130815,
		131865,
		131299,
		42009,
		130839,
		131386,
		130719,
		131038,
		131367,
		130715,
		131693,
		131317,
		130707,
		130709,
		131691,
		131379,
		131663,
		131298,
		131666,
		131493,
		131374,
		131360,
		131535,
		131634,
		131051,
		131380,
		131363,
		42005,
		130792,
		131362,
		131387,
		131364,
		130661,
		131519,
		42013,
		132034,
		132033,
		131417,
		131375,
		131492,
		131528,
		131356,
		131397,
		131662,
		131115,
		131899,
		131353,
		131287,
		131671,
		131490,
		131344,
		131590,
		130082,
		130642,
		131325,
		130842,
		131305,
		131355,
		132077,
		131354,
		131304,
		131462,
		131999,
		131548,
		131286,
		131659,
		130843,
		131649,
		131518,
		131390,
		131330,
		131357,
		42019,
		131365,
		132005,
		131521,
		131549,
		131361,
		131382,
		131323,
		131889,
		42011,
		131591,
		131463,
		130714,
		131544,
		131680,
		131385,
		130894,
		131443,
		131670,
		130793,
		131407,
		131837,
		42023,
		131592,
		131331,
		130200,
		131346,
		131525,
		42021,
		131633,
		130998,
		130660,
		131122,
		132070,
		131998,
		131381,
		131383,
		131366,
		130907,
		131023,
		131546,
		130136,
		131841,
		132017,
		132031,
		131545,
		131450,
		131669,
		131124,
		132030,
		131441,
		131118,
		132071,
		131338,
		130829,
		131888,
		131524,
		131949,
		131401,
		131690,
		131351,
		130017,
		130069,
		131316,
		132004,
		131491,
		131419,
		131498,
		130019,
		131499,
		131049,
		131529,
		130345,
		130990,
		131104,
		132007,
		131919,
		131901,
		131543,
		132280,
		131036,
		131652,
		132035,
		131625,
		131352,
		130830,
		130274,
		131667,
		131488,
		130648,
		131017,
		130008,
		131121,
		132032,
		131001,
		131378,
		130828,
		131675,
		132481,
		130659,
		131834,
		130862,
		131876,
		131345,
		131930,
		131648,
		131683,
		131603,
		131335,
		130895,
		131635,
		131636,
		132469,
		131276,
		130854,
		130855,
		130018,
		131624,
		131630,
		132092,
		131048,
		131914,
		130686,
		132058,
		131547,
		131593,
		130083,
		131065,
		132093,
		131864,
		130826,
		130344,
		131339,
		131689,
		130068,
		130992,
		130197,
		131595,
		130846,
		131514,
		131971,
		131109,
		130688,
		131114,
		131284,
		131668,
		130178,
		131714,
		131322,
		131965,
		131848,
		131637,
		130206,
		131000,
		131277,
		130255,
		131661,
		130708,
		131586,
		131910,
		131400,
		131654,
		130729,
		131311,
		130838,
		131384,
		132055,
		130658,
		131709,
		131532,
		130137,
		131107,
		131046,
		131125,
		131402,
		131587,
		131321,
		131039,
		130825,
		131505,
		130814,
		130906,
		131684,
		130869,
		131406,
		130215,
		131011,
		131391,
		132057,
		131422,
		131503,
		131523,
		131653,
		132472,
		132513,
		131016,
		42025,
		131609,
		130009,
		130205,
		130831,
		131110,
		130827,
		131119,
		132023,
		131111,
		131022,
		131655,
		131291,
		131415,
		130747,
		130824,
		130212,
		131681,
		131931,
		131108,
		130330,
		130786,
		131631,
		130666,
		131702,
		131538,
		132020,
		131483,
		130016,
		131533,
		132006,
		130689,
		131320,
		132051,
		132104,
		131970,
		131388,
		131612,
		131315,
		2443,
		42003,
		3079,
		131929,
		131955,
		131229,
		130868,
		132123,
		131946,
		131142,
		131223,
		131157,
		131495,
		131928,
		131768,
		131774,
		131328,
		131222,
		132140,
		131917,
		131235,
		131938,
		131464,
		131225,
		132152,
		131701,
		132163,
		131143,
		131721,
		131939,
		130226,
		131807,
		131187,
		131256,
		131738,
		131961,
		131258,
		131986,
		131754,
		131878,
		131430,
		131349,
		131769,
		132141,
		131060,
		131650,
		130910,
		131467,
		131465,
		131191,
		131700,
		130188,
		131054,
		131584,
		131782,
		131836,
		132428,
		132062,
		130189,
		131437,
		131057,
		131494,
		130176,
		131916,
		131055,
		130218,
		131960,
		131790,
		131056,
		131263,
		131616,
		132096,
		131954,
		131130,
		131651,
		131613,
		131076,
		131859,
		131451,
		132151,
		131429,
		132103,
		130737,
		131741,
		131766,
		131892,
		131956,
		131755,
		131253,
		130179,
		131617,
		131872,
		130224,
		132277,
		131639,
		131870,
		131951,
		132476,
		131676,
		131297,
		131190,
		131446,
		131262,
		132153,
		131427,
		131877,
		132072,
		131794,
		131265,
		130227,
		131793,
		131561,
		131759,
		131137,
		132125,
		130915,
		131217,
		131033,
		130777,
		130897,
		131216,
		131471,
		132105,
		131425,
		130622,
		130873,
		131466,
		131252,
		131136,
		131957,
		132119,
		131053,
		131640,
		132066,
		131950,
		131943,
		131074,
		130320,
		131436,
		131885,
		131708,
		131426,
		132220,
		131959,
		130105,
		132157,
		131210,
		131224,
		131972,
		131261,
		132530,
		131958,
		131806,
		131737,
		131791,
		132144,
		131720,
		131740,
		131469,
		131416,
		131144,
		130784,
		132150,
		131052,
		132097,
		131209,
		132128,
		131883,
		132146,
		131948,
		131858,
		131851,
		132073,
		131934,
		131810,
		131264,
		132063,
		130967,
		130106,
		131925,
		131585,
		131703,
		131991,
		131442,
		132169,
		131230,
		130896,
		130911,
		130623,
		131032,
		131581,
		130965,
		131723,
		132133,
		130175,
		130860,
		131811,
		131432,
		132043,
		131138,
		130820,
		131063,
		131484,
		131182,
		131722,
		132054,
		130310,
		131869,
		131911,
		132145,
		131166,
		132102,
		131743,
		130231,
		131646,
		131424,
		130023,
		132113,
		130174,
		131468,
		132137,
		131283,
		131128,
		132131,
		131976,
		131906,
		131482,
		131871,
		131734,
		131106,
		130718,
		131431,
		130228,
		131348,
		130807,
		131739,
		131775,
		132175,
		130922,
		131918,
		131747,
		132281,
		131825,
		131698,
		131403,
		132156,
		131879,
		130062,
		130771,
		131145,
		131459,
		131778,
		131873,
		131647,
		132203,
		131186,
		132160,
		131234,
		132036,
		132602,
		132147,
		131599,
		131428,
		131204,
		131243,
		131935,
		132465,
		132167,
		131472,
		132136,
		131132,
		131742,
		131075,
		130904,
		131600,
		131473,
		131621,
		131458,
		131205,
		131058,
		131576,
		131861,
		131259,
		131643,
		132089,
		130776,
		131579,
		131638,
		130015,
		131061,
		130909,
		131981,
		131237,
		132149,
		130914,
		131926,
		131148,
		131231,
		131236,
		131129,
		131993,
		132276,
		131824,
		130802,
		130185,
		131767,
		132142,
		130321,
		132168,
		131641,
		132190,
		130184,
		131560,
		131477,
		132192,
		131183,
		131260,
		130657,
		132118,
		131062,
		131792,
		131573,
		131605,
		131797,
		131133,
		130122,
		131796,
		130693,
		131077,
		131922,
		131736,
		131820,
		131947,
		131598,
		132161,
		131927,
		131614,
		132124,
		130229,
		131942,
		132138,
		131724,
		131257,
		130219,
		132101,
		132117,
		131868,
		131688,
		130230,
		131392,
		130989,
		131064,
		130699,
		132100,
		131880,
		130893,
		130905,
		130225,
		131131,
		130902,
		130919,
		132501,
		132099,
		131679,
		131413,
		131571,
		130670,
		131804,
		131178,
		131192,
		131280,
		131193,
		131578,
		131764,
		132088,
		130861,
		130177,
		131070,
		131554,
		131043,
		131973,
		130625,
		131557,
		131577,
		132083,
		131813,
		130698,
		131988,
		131159,
		132538,
		130687,
		131470,
		131812,
		131760,
		131964,
		131992,
		130026,
		131904,
		131886,
		131583,
		131408,
		130908,
		132518,
		131566,
		130770,
		130900,
		132059,
		131882,
		131150,
		130194,
		131574,
		130890,
		131158,
		130920,
		132195,
		131795,
		131582,
		131789,
		132148,
		131731,
		131272,
		131615,
		130736,
		131980,
		131678,
		131730,
		131208,
		131745,
		130005,
		130626,
		131139,
		132279,
		132143,
		132278,
		130092,
		132042,
		131180,
		131924,
		130944,
		131618,
		132019,
		131042,
		130325,
		131627,
		131990,
		130818,
		130705,
		131228,
		131414,
		132225,
		131212,
		131827,
		132056,
		130921,
		130671,
		130898,
		131153,
		131154,
		130901,
		132027,
		131580,
		131433,
		131628,
		131149,
		131485,
		131066,
		132264,
		132129,
		130678,
		130216,
		131165,
		131823,
		132540,
		131181,
		131941,
		130888,
		131412,
		131642,
		132542,
		131748,
		130649,
		131155,
		130923,
		131601,
		132116,
		131565,
		131783,
		130217,
		131604,
		131296,
		130988,
		132520,
		132021,
		132197,
		131213,
		132215,
		131699,
		130697,
		130066,
		130063,
		130924,
		131575,
		131242,
		131761,
		132545,
		131059,
		130891,
		131175,
		132171,
		132134,
		131244,
		130949,
		132085,
		131900,
		130809,
		131989,
		131704,
		131987,
		130186,
		131897,
		132311,
		130635,
		131211,
		131835,
		131152,
		130190,
		131198,
		131167,
		131540,
		131447,
		132291,
		132575,
		131978,
		131803,
		132112,
		131620,
		130191,
		131393,
		131086,
		131251,
		131779,
		132563,
		130819,
		132460,
		132184,
		132537,
		131674,
		130775,
		132267,
		131726,
		130004,
		131705,
		130254,
		130318,
		132170,
		131247,
		130805,
		131480,
		131974,
		131915,
		132130,
		130633,
		130917,
		130945,
		132298,
		130849,
		132456,
		132080,
		132265,
		131788,
		130717,
		132135,
		131481,
		131079,
		132037,
		131770,
		132202,
		130725,
		132016,
		131677,
		132109,
		131896,
		131644,
		131572,
		130696,
		130963,
		132431,
		130889,
		131197,
		131541,
		131746,
		131478,
		130195,
		132082,
		130712,
		130093,
		131799,
		131151,
		130701,
		131246,
		131729,
		132574,
		131029,
		132078,
		132091,
		132271,
		131031,
		130281,
		130119,
		132461,
		131567,
		131975,
		132194,
		131784,
		131757,
		132098,
		132087,
		131207,
		132084,
		132075,
		130656,
		131887,
		132081,
		132174,
		131334,
		130872,
		131645,
		131822,
		132365,
		131907,
		131940,
		131103,
		130223,
		132527,
		131409,
		130107,
		131849,
		132468,
		132067,
		131594,
		131411,
		130700,
		131281,
		130022,
		132110,
		131275,
		130643,
		132474,
		132166,
		130010,
		132074,
		130664,
		131082,
		131179,
		132541,
		132443,
		130833,
		131805,
		131826,
		132039,
		131781,
		131435,
		131749,
		130875,
		131619,
		132189,
		131342,
		131021,
		132196,
		131526,
		131898,
		132299,
		130677,
		131985,
		131751,
		132003,
		130779,
		132533,
		132090,
		130148,
		130196,
		130726,
		132268,
		132543,
		130187,
		130823,
		132320,
		131047,
		131863,
		132079,
		130916,
		130844,
		131685,
		132269,
		131765,
		131707,
		131562,
		131798,
		132139,
		131909,
		130311,
		131094,
		130733,
		131371,
		131850,
		132587,
		131881,
		131093,
		132217,
		131563,
		132442,
		132536,
		132064,
		132578,
		130722,
		132045,
		130852,
		131201,
		131808,
		132462,
		131453,
		131558,
		130096,
		131894,
		132270,
		130118,
		131445,
		132544,
		130240,
		131623,
		132044,
		131410,
		130774,
		132455,
		130324,
		130903,
		130918,
		131087,
		131844,
		131821,
		131295,
		131923,
		132111,
		132108,
		131710,
		131711,
		130202,
		131862,
		130727,
		130341,
		130647,
		130723,
		132201,
		130013,
		130014,
		132212,
		132457,
		130821,
		131884,
		131860,
		130287,
		130293,
		132132,
		132313,
		132531,
		132532,
		132275,
		131440,
		131758,
		132341,
		130249,
		130787,
		131962,
		131556,
		131239,
		131294,
		131097,
		132521,
		132585,
		131489,
		131913,
		131161,
		132340,
		132580,
		131785,
		130785,
		130816,
		131780,
		132312,
		131750,
		132570,
		130746,
		132505,
		132162,
		130624,
		132523,
		132576,
		132586,
		130912,
		131398,
		132188,
		132094,
		130059,
		132213,
		132430,
		131522,
		130102,
		130104,
		131569,
		131728,
		132106,
		131071,
		132504,
		132002,
		131564,
		132490,
		130806,
		131174,
		131610,
		131802,
		132566,
		132519,
		132559,
		130817,
		131984,
		131727,
		131343,
		131101,
		130164,
		131214,
		131202,
		130319,
		131083,
		132475,
		131474,
		131905,
		131250,
		131196,
		131629,
		131078,
		132526,
		130003,
		131487,
		132038,
		131030,
		130704,
		131713,
		132547,
		131515,
		130149,
		132122,
		131496,
		130046,
		131461,
		130879,
		132009,
		132193,
		131240,
		132562,
		130798,
		131706,
		131893,
		131570,
		132018,
		130951,
		131314,
		131475,
		132026,
		130253,
		131476,
		130676,
		131875,
		130203,
		131626,
		130058,
		131160,
		131096,
		131539,
		130161,
		132282,
		132569,
		131830,
		131744,
		130064,
		132516,
		131479,
		132300,
		131067,
		130881,
		132454,
		132107,
		132571,
		130337,
		130665,
		130703,
		130962,
		130065,
		130721,
		132076,
		131542,
		131602,
		132086,
		131164,
		130702,
		131084,
		131014,
		130993,
		130738,
		130627,
		132429,
		130808,
		130724,
		131199,
		131735,
		131979,
		130913,
		130072,
		131173,
		130948,
		132301,
		132244,
		132321,
		132200,
		130011,
		130315,
		132511,
		130899,
		130850,
		132482,
		130198,
		132173,
		132065,
		131172,
		130692,
		132584,
		132022,
		130739,
		130782,
		132380,
		132500,
		131996,
		130634,
		130336,
		132535,
		132491,
		132581,
		132199,
		130073,
		132453,
		132290,
		131840,
		132370,
		130264,
		130043,
		130694,
		130220,
		130028,
		131245,
		130885,
		131005,
		130027,
		130150,
		130222,
		131756,
		131502,
		132480,
		132266,
		130925,
		131278,
		132221,
		132579,
		132296,
		132226,
		130097,
		132191,
		131227,
		130629,
		132449,
		130982,
		130695,
		131203,
		131818,
		131665,
		131534,
		131717,
		132525,
		130884,
		132558,
		130740,
		130639,
		132452,
		132448,
		131568,
		130214,
		131912,
		130078,
		131831,
		130103,
		131486,
		130735,
		130932,
		130638,
		132477,
		130933,
		132451,
		131102,
		131206,
		130874,
		130124,
		130047,
		132245,
		130892,
		130280,
		130804,
		130192,
		130937,
		131833,
		132186,
		130252,
		131147,
		130159,
		131266,
		132564,
		130979,
		131842,
		131452,
		130012,
		132484,
		130006,
		132224,
		132553,
		130865,
		130720,
		130123,
		132120,
		132187,
		130848,
		131434,
		130067,
		130685,
		130168,
		131664,
		131418,
		130221,
		131215,
		131520,
		130791,
		131327,
		131120,
		130128,
		130970,
		132502,
		130832,
		132489,
		132181,
		130130,
		131660,
		130706,
		130314,
		131497,
		132483,
		132391,
		132207,
		130732,
		131607,
		130628,
		131553,
		132510,
		131995,
		131073,
		130037,
		132216,
		131828,
		132517,
		131832,
		131135,
		132287,
		132464,
		130728,
		130864,
		132121,
		131597,
		132598,
		130636,
		130997,
		131072,
		130632,
		131809,
		130262,
		132512,
		131874,
		132577,
		130667,
		130940,
		130856,
		130129,
		130731,
		131845,
		131241,
		132438,
		132208,
		130338,
		131867,
		131095,
		131932,
		130138,
		130141,
		131146,
		131176,
		130663,
		130086,
		131290,
		131963,
		131232,
		130266,
		130256,
		132573,
		131282,
		130201,
		132047,
		130339,
		130762,
		131015,
		132568,
		132522,
		131555,
		131085,
		132362,
		131997,
		130199,
		131092,
		131829,
		132604,
		131189,
		130060,
		130248,
		131527,
		130964,
		131552,
		132447,
		130309,
		130079,
		132049,
		130113,
		131040,
		132178,
		130081,
		130001,
		132127,
		131025,
		132503,
		131279,
		131606,
		130822,
		131238,
		132198,
		130272,
		130978,
		130303,
		130158,
		130971,
		130265,
		132473,
		131169,
		130131,
		132554,
		130153,
		132364,
		130851,
		131270,
		131050,
		130942,
		131271,
		132589,
		130151,
		130880,
		130684,
		132485,
		132390,
		130866,
		131611,
		132179,
		132567,
		130259,
		130679,
		130134,
		130080,
		132594,
		130783,
		132446,
		131506,
		132605,
		130934,
		130257,
		131002,
		132551,
		131269,
		132561,
		131725,
		131389,
		131945,
		131100,
		131507,
		131123,
		132288,
		130646,
		132238,
		130091,
		130051,
		130316,
		132546,
		132445,
		131018,
		131855,
		132360,
		131977,
		132369,
		130180,
		132552,
		130980,
		130273,
		130878,
		131329,
		130952,
		131801,
		132048,
		130007,
		131218,
		131444,
		132539,
		131024,
		130258,
		130637,
		132549,
		131786,
		131004,
		130087,
		132172,
		131009,
		130991,
		130857,
		132206,
		132293,
		132444,
		132114,
		132158,
		130803,
		131715,
		130306,
		132534,
		131967,
		130983,
		130859,
		130730,
		131673,
		131697,
		130645,
		132463,
		130716,
		130938,
		131267,
		130245,
		130977,
		131969,
		131773,
		131460,
		130976,
		130084,
		131712,
		132486,
		132398,
		130960,
		132246,
		131370,
		131908,
		132185,
		130999,
		132416,
		130244,
		131301,
		131373,
		131003,
		130734,
		132283,
		132297,
		130169,
		132383,
		131156,
		130135,
		131787,
		130973,
		131983,
		132560,
		130683,
		132164,
		130758,
		132314,
		131816,
		130165,
		132350,
		130193,
		130054,
		132180,
		132599,
		131933,
		132223,
		132433,
		132240,
		132008,
		131105,
		132499,
		132471,
		131843,
		132237,
		132227,
		132165,
		130317,
		132235,
		132550,
		132126,
		130662,
		131890,
		130002,
		132257,
		130305,
		130741,
		131249,
		130030,
		130121,
		132515,
		131508,
		132487,
		132374,
		131008,
		132458,
		130312,
		132422,
		131716,
		130795,
		131982,
		131326,
		132470,
		132286,
		130713,
		131622,
		132319,
		130307,
		132338,
		130943,
		131455,
		130152,
		132583,
		132432,
		130966,
		131134,
		132524,
		131177,
		130302,
		131866,
		132498,
		130050,
		130340,
		131672,
		132479,
		130631,
		130790,
		130858,
		131772,
		130941,
		132025,
		132326,
		131920,
		131313,
		131285,
		131127,
		132334,
		131994,
		132466,
		131819,
		131531,
		131168,
		131454,
		130763,
		131530,
		130994,
		130987,
		130095,
		130752,
		131195,
		131200,
		130213,
		132307,
		132397,
		132014,
		130061,
		132011,
		132289,
		132459,
		130813,
		132603,
		132597,
		132274,
		132592,
		130653,
		132347,
		131226,
		132209,
		131596,
		130847,
		130996,
		130044,
		131516,
		132508,
		132239,
		130139,
		132434,
		130120,
		131944,
		132441,
		132183,
		130286,
		131081,
		131012,
		131895,
		130929,
		132337,
		132335,
		130887,
		131027,
		132310,
		130654,
		132351,
		130650,
		132572,
		132596,
		132327,
		130672,
		130000,
		132488,
		131044,
		131509,
		130045,
		132440,
		130112,
		131068,
		130147,
		130090,
		130995,
		131732,
		130145,
		130778,
		131800,
		130870,
		132258,
		132371,
		132115,
		131293,
		132588,
		130282,
		130297,
		130742,
		130285,
		131852,
		130682,
		130810,
		131372,
		131013,
		131405,
		132261,
		131404,
		131184,
		130343,
		131300,
		132250,
		132241,
		132231,
		130840,
		131853,
		131449,
		132249,
		132262,
		132230,
		132256,
		132260,
		132242,
		131312,
		132263,
		131420,
		131421,
		131141,
		132248,
		132229,
		131221,
		132251,
		131185,
		130342,
		132228,
		130275,
		131098,
		131902,
		131088,
		131117,
		131966,
		132233,
		131537,
		131089,
		130837,
		130836,
		131163,
		131126,
		131608,
		130021,
		131399,
		130841,
		132001,
		130328,
		130644,
		132015,
		131921,
		130935,
		131536,
		132243,
		132253,
		131589,
		131551,
		130210,
		132234,
		131332,
		131846,
		131395,
		131248,
		131020,
		131220,
		132222,
		131500,
		131694,
		132232,
		131010,
		131041,
		131080,
		130766,
		130655,
		132214,
		131140,
		132219,
		131657,
		130640,
		132254,
		131448,
		132252,
		132046,
		131268,
		131302,
		132155,
		131550,
		132029,
		132272,
		132154,
		130331,
		131559,
		131219,
		132218,
		130673,
		131069,
		131162,
		132010,
		130950,
		130089,
		130146,
		132095,
		130961,
		130710,
		131394,
		132591,
		132024,
		132514,
		131099,
		132285,
		131588,
		130959,
		130042,
		130744,
		132236,
		130334,
		130088,
		132247,
		132052,
		132435,
		130772,
		132259,
		130799,
		131116,
		131854,
		131838,
		131045,
		132294,
		131292,
		130233,
		130652,
		130745,
		131319,
		132555,
		132255,
		130886,
		130036,
		131695,
		132565,
		130743,
		132439,
		131968,
		131501,
		132292,
		130680,
		131817,
		132176,
		130985,
		132000,
		131340,
		131090,
		130764,
		130947,
		130335,
		130756,
		132295,
		130025,
		131170,
		130020,
		130211,
		131903,
		132182,
		130867,
		132601,
		130877,
		131376,
		131113,
		130277,
		130681,
		132557,
		130247,
		130284,
		130757,
		130116,
		130928,
		130669,
		130773,
		132467,
		132495,
		132478,
		131341,
		132450,
		130276,
		132600,
		131377,
		131171,
		131771,
		130765,
		132068,
		130630,
		130204,
		130142,
		130251,
		130863,
		130024,
		130801,
		131303,
		132366,
		130711,
		132028,
		130845,
		132492,
		132273,
		130939,
		130767,
		130853,
		130304,
		132349,
		130250,
		130797,
		131358,
		130155,
		130986,
		132060,
		132159,
		132053,
		130207,
		130936,
		132210,
		131194,
		130789,
		131019,
		130811,
		132497,
		131438,
		131504,
		130132,
		130329,
		132494,
		131273,
		130975,
		130327,
		130641,
		130246,
		132177,
		130053,
		130242,
		132323,
		132284,
		131091,
		131439,
		131359,
		130170,
		132361,
		132358,
		130313,
		132318,
		130289,
		132359,
		130984,
		130114,
		130085,
		132336,
		132342,
		131274,
		130181,
		131233,
		132506,
		130140,
		132427,
		130163,
		131839,
		132493,
		132306,
		132423,
		132353,
		132382,
		132315,
		132346,
		132388,
		130931,
		132389,
		130871,
		130263,
		132330,
		131517,
		132381,
		132363,
		130160,
		132329,
		130029,
		130296,
		130162,
		132529,
		132509,
		132332,
		132395,
		130292,
		132378,
		132375,
		130288,
		131733,
		130234,
		130972,
		132528,
		130781,
		130946,
		130323,
		132496,
		132339,
		132426,
		132012,
		130074,
		130267,
		130955,
		130260,
		130209,
		130927,
		131318,
		131719,
		131337,
		130270,
		130876,
		131847,
		132590,
		132303,
		132013,
		130208,
		130759,
		132394,
		131112,
		132548,
		132348,
		132205,
		130668,
		130133,
		130974,
		132595,
		130926,
		131513,
		130674,
		130800,
		132507,
		132069,
		130794,
		131336,
		130154,
		131333,
		130812,
		130125,
		130981,
		130031,
		132377,
		132343,
		130261,
		132367,
		130283,
		130298,
		130953,
		130753,
		132368,
		132417,
		130322,
		132356,
		132328,
		132344,
		130075,
		130052,
		132302,
		130035,
		131656,
		130748,
		132593,
		131687,
		130235,
		130100,
		131006,
		130127,
		131696,
		132061,
		132436,
		130769,
		130271,
		130144,
		130109,
		132333,
		130954,
		132556,
		131776,
		130034,
		130143,
		130041,
		132410,
		132345,
		132376,
		130101,
		131814,
		130241,
		132324,
		132325,
		130094,
		132040,
		131007,
		130958,
		131686,
		132420,
		130882,
		130760,
		132582,
		130108,
		130883,
		130156,
		131762,
		131188,
		131952,
		130126,
		132396,
		132331,
		132211,
		130117,
		130171,
		130049,
		131026,
		130308,
		130326,
		130768,
		132421,
		132411,
		130780,
		130299,
		130076,
		131936,
		132204,
		130749,
		130930,
		131512,
		130237,
		132437,
		130301,
		130115,
		130243,
		131937,
		131763,
		130788,
		130761,
		132352,
		132355,
		132379,
		132354,
		132372,
		130040,
		131777,
		130300,
		130048,
		132399,
		130675,
		131718,
		130157,
		130038,
		130232,
		131255,
		131815,
		132373,
		130796,
		130077,
		130055,
		132357,
		131254,
		132322,
		132041,
		130236,
		130039,
		131953,
	}

	// Add all of these as known FieldIDs
	l0 := len(MaterializedSpecial)
	for _, v := range knownSpecialFields {
		MaterializedSpecial[v] = struct{}{}
	}
	l1 := len(MaterializedSpecial)

	log.Printf("Observed %d special fields with known dates (%d total, including the %d statically coded fields)", l1-l0, l1, l0)

}
