package locale

import (
	"strings"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/registry"
	"github.com/noxyicm/wsf/utils"
)

var (
	// List of locales that are no longer part of CLDR along with a
	// mapping to an appropriate alternative
	localeAliases = map[string]string{
		"az_AZ":  "az_Latn_AZ",
		"bs_BA":  "bs_Latn_BA",
		"ha_GH":  "ha_Latn_GH",
		"ha_NE":  "ha_Latn_NE",
		"ha_NG":  "ha_Latn_NG",
		"kk_KZ":  "kk_Cyrl_KZ",
		"ks_IN":  "ks_Arab_IN",
		"mn_MN":  "mn_Cyrl_MN",
		"ms_BN":  "ms_Latn_BN",
		"ms_MY":  "ms_Latn_MY",
		"ms_SG":  "ms_Latn_SG",
		"pa_IN":  "pa_Guru_IN",
		"pa_PK":  "pa_Arab_PK",
		"shi_MA": "shi_Latn_MA",
		"sr_BA":  "sr_Latn_BA",
		"sr_ME":  "sr_Latn_ME",
		"sr_RS":  "sr_Latn_RS",
		"sr_XK":  "sr_Latn_XK",
		"tg_TJ":  "tg_Cyrl_TJ",
		"tzm_MA": "tzm_Latn_MA",
		"uz_AF":  "uz_Arab_AF",
		"uz_UZ":  "uz_Latn_UZ",
		"vai_LR": "vai_Latn_LR",
		"zh_CN":  "zh_Hans_CN",
		"zh_HK":  "zh_Hant_HK",
		"zh_MO":  "zh_Hans_MO",
		"zh_SG":  "zh_Hans_SG",
		"zh_TW":  "zh_Hant_TW",
	}

	// Class wide Locale Constants
	localeData = map[string]bool{
		"root":           true,
		"aa":             true,
		"aa_DJ":          true,
		"aa_ER":          true,
		"aa_ET":          true,
		"af":             true,
		"af_NA":          true,
		"af_ZA":          true,
		"agq":            true,
		"agq_CM":         true,
		"ak":             true,
		"ak_GH":          true,
		"am":             true,
		"am_ET":          true,
		"ar":             true,
		"ar_001":         true,
		"ar_AE":          true,
		"ar_BH":          true,
		"ar_DJ":          true,
		"ar_DZ":          true,
		"ar_EG":          true,
		"ar_EH":          true,
		"ar_ER":          true,
		"ar_IL":          true,
		"ar_IQ":          true,
		"ar_JO":          true,
		"ar_KM":          true,
		"ar_KW":          true,
		"ar_LB":          true,
		"ar_LY":          true,
		"ar_MA":          true,
		"ar_MR":          true,
		"ar_OM":          true,
		"ar_PS":          true,
		"ar_QA":          true,
		"ar_SA":          true,
		"ar_SD":          true,
		"ar_SO":          true,
		"ar_SS":          true,
		"ar_SY":          true,
		"ar_TD":          true,
		"ar_TN":          true,
		"ar_YE":          true,
		"as":             true,
		"as_IN":          true,
		"asa":            true,
		"asa_TZ":         true,
		"ast":            true,
		"ast_ES":         true,
		"az":             true,
		"az_Cyrl":        true,
		"az_Cyrl_AZ":     true,
		"az_Latn":        true,
		"az_Latn_AZ":     true,
		"bas":            true,
		"bas_CM":         true,
		"be":             true,
		"be_BY":          true,
		"bem":            true,
		"bem_ZM":         true,
		"bez":            true,
		"bez_TZ":         true,
		"bg":             true,
		"bg_BG":          true,
		"bm":             true,
		"bm_ML":          true,
		"bn":             true,
		"bn_BD":          true,
		"bn_IN":          true,
		"bo":             true,
		"bo_CN":          true,
		"bo_IN":          true,
		"br":             true,
		"br_FR":          true,
		"brx":            true,
		"brx_IN":         true,
		"bs":             true,
		"bs_Cyrl":        true,
		"bs_Cyrl_BA":     true,
		"bs_Latn":        true,
		"bs_Latn_BA":     true,
		"byn":            true,
		"byn_ER":         true,
		"ca":             true,
		"ca_AD":          true,
		"ca_ES":          true,
		"ca_ES_VALENCIA": true,
		"ca_FR":          true,
		"ca_IT":          true,
		"cgg":            true,
		"cgg_UG":         true,
		"chr":            true,
		"chr_US":         true,
		"cs":             true,
		"cs_CZ":          true,
		"cy":             true,
		"cy_GB":          true,
		"da":             true,
		"da_DK":          true,
		"da_GL":          true,
		"dav":            true,
		"dav_KE":         true,
		"de":             true,
		"de_AT":          true,
		"de_BE":          true,
		"de_CH":          true,
		"de_DE":          true,
		"de_LI":          true,
		"de_LU":          true,
		"dje":            true,
		"dje_NE":         true,
		"dua":            true,
		"dua_CM":         true,
		"dyo":            true,
		"dyo_SN":         true,
		"dz":             true,
		"dz_BT":          true,
		"ebu":            true,
		"ebu_KE":         true,
		"ee":             true,
		"ee_GH":          true,
		"ee_TG":          true,
		"el":             true,
		"el_CY":          true,
		"el_GR":          true,
		"en":             true,
		"en_001":         true,
		"en_150":         true,
		"en_AG":          true,
		"en_AI":          true,
		"en_AS":          true,
		"en_AU":          true,
		"en_BB":          true,
		"en_BE":          true,
		"en_BM":          true,
		"en_BS":          true,
		"en_BW":          true,
		"en_BZ":          true,
		"en_CA":          true,
		"en_CC":          true,
		"en_CK":          true,
		"en_CM":          true,
		"en_CX":          true,
		"en_DG":          true,
		"en_DM":          true,
		"en_Dsrt":        true,
		"en_Dsrt_US":     true,
		"en_ER":          true,
		"en_FJ":          true,
		"en_FK":          true,
		"en_FM":          true,
		"en_GB":          true,
		"en_GD":          true,
		"en_GG":          true,
		"en_GH":          true,
		"en_GI":          true,
		"en_GM":          true,
		"en_GU":          true,
		"en_GY":          true,
		"en_HK":          true,
		"en_IE":          true,
		"en_IM":          true,
		"en_IN":          true,
		"en_IO":          true,
		"en_JE":          true,
		"en_JM":          true,
		"en_KE":          true,
		"en_KI":          true,
		"en_KN":          true,
		"en_KY":          true,
		"en_LC":          true,
		"en_LR":          true,
		"en_LS":          true,
		"en_MG":          true,
		"en_MH":          true,
		"en_MO":          true,
		"en_MP":          true,
		"en_MS":          true,
		"en_MT":          true,
		"en_MU":          true,
		"en_MW":          true,
		"en_NA":          true,
		"en_NF":          true,
		"en_NG":          true,
		"en_NR":          true,
		"en_NU":          true,
		"en_NZ":          true,
		"en_PG":          true,
		"en_PH":          true,
		"en_PK":          true,
		"en_PN":          true,
		"en_PR":          true,
		"en_PW":          true,
		"en_RW":          true,
		"en_SB":          true,
		"en_SC":          true,
		"en_SD":          true,
		"en_SG":          true,
		"en_SH":          true,
		"en_SL":          true,
		"en_SS":          true,
		"en_SX":          true,
		"en_SZ":          true,
		"en_TC":          true,
		"en_TK":          true,
		"en_TO":          true,
		"en_TT":          true,
		"en_TV":          true,
		"en_TZ":          true,
		"en_UG":          true,
		"en_UM":          true,
		"en_US":          true,
		"en_US_POSIX":    true,
		"en_VC":          true,
		"en_VG":          true,
		"en_VI":          true,
		"en_VU":          true,
		"en_WS":          true,
		"en_ZA":          true,
		"en_ZM":          true,
		"en_ZW":          true,
		"eo":             true,
		"eo_001":         true,
		"es":             true,
		"es_419":         true,
		"es_AR":          true,
		"es_BO":          true,
		"es_CL":          true,
		"es_CO":          true,
		"es_CR":          true,
		"es_CU":          true,
		"es_DO":          true,
		"es_EA":          true,
		"es_EC":          true,
		"es_ES":          true,
		"es_GQ":          true,
		"es_GT":          true,
		"es_HN":          true,
		"es_IC":          true,
		"es_MX":          true,
		"es_NI":          true,
		"es_PA":          true,
		"es_PE":          true,
		"es_PH":          true,
		"es_PR":          true,
		"es_PY":          true,
		"es_SV":          true,
		"es_US":          true,
		"es_UY":          true,
		"es_VE":          true,
		"et":             true,
		"et_EE":          true,
		"eu":             true,
		"eu_ES":          true,
		"ewo":            true,
		"ewo_CM":         true,
		"fa":             true,
		"fa_AF":          true,
		"fa_IR":          true,
		"ff":             true,
		"ff_CM":          true,
		"ff_GN":          true,
		"ff_MR":          true,
		"fr_PM":          true,
		"ff_SN":          true,
		"fr_WF":          true,
		"fi":             true,
		"fi_FI":          true,
		"fil":            true,
		"fil_PH":         true,
		"fo":             true,
		"fo_FO":          true,
		"fr":             true,
		"fr_BE":          true,
		"fr_BF":          true,
		"fr_BI":          true,
		"fr_BJ":          true,
		"fr_BL":          true,
		"fr_CA":          true,
		"fr_CD":          true,
		"fr_CF":          true,
		"fr_CG":          true,
		"fr_CH":          true,
		"fr_CI":          true,
		"fr_CM":          true,
		"fr_DJ":          true,
		"fr_DZ":          true,
		"fr_FR":          true,
		"fr_GA":          true,
		"fr_GF":          true,
		"fr_GN":          true,
		"fr_GP":          true,
		"fr_GQ":          true,
		"fr_HT":          true,
		"fr_KM":          true,
		"fr_LU":          true,
		"fr_MA":          true,
		"fr_MC":          true,
		"fr_MF":          true,
		"fr_MG":          true,
		"fr_ML":          true,
		"fr_MQ":          true,
		"fr_MR":          true,
		"fr_MU":          true,
		"fr_NC":          true,
		"fr_NE":          true,
		"fr_PF":          true,
		"fr_RE":          true,
		"fr_RW":          true,
		"fr_SC":          true,
		"fr_SN":          true,
		"fr_SY":          true,
		"fr_TD":          true,
		"fr_TG":          true,
		"fr_TN":          true,
		"fr_VU":          true,
		"fr_YT":          true,
		"fur":            true,
		"fur_IT":         true,
		"fy":             true,
		"fy_NL":          true,
		"ga":             true,
		"ga_IE":          true,
		"gd":             true,
		"gd_GB":          true,
		"gl":             true,
		"gl_ES":          true,
		"gsw":            true,
		"gsw_CH":         true,
		"gsw_LI":         true,
		"gu":             true,
		"gu_IN":          true,
		"guz":            true,
		"guz_KE":         true,
		"gv":             true,
		"gv_IM":          true,
		"ha":             true,
		"ha_Latn":        true,
		"ha_Latn_GH":     true,
		"ha_Latn_NE":     true,
		"ha_Latn_NG":     true,
		"haw":            true,
		"haw_US":         true,
		"he":             true,
		"he_IL":          true,
		"hi":             true,
		"hi_IN":          true,
		"hr":             true,
		"hr_BA":          true,
		"hr_HR":          true,
		"hu":             true,
		"hu_HU":          true,
		"hy":             true,
		"hy_AM":          true,
		"ia":             true,
		"ia_FR":          true,
		"id":             true,
		"id_ID":          true,
		"ig":             true,
		"ig_NG":          true,
		"ii":             true,
		"ii_CN":          true,
		"is":             true,
		"is_IS":          true,
		"it":             true,
		"it_CH":          true,
		"it_IT":          true,
		"it_SM":          true,
		"ja":             true,
		"ja_JP":          true,
		"jgo":            true,
		"jgo_CM":         true,
		"jmc":            true,
		"jmc_TZ":         true,
		"ka":             true,
		"ka_GE":          true,
		"kab":            true,
		"kab_DZ":         true,
		"kam":            true,
		"kam_KE":         true,
		"kde":            true,
		"kde_TZ":         true,
		"kea":            true,
		"kea_CV":         true,
		"khq":            true,
		"khq_ML":         true,
		"ki":             true,
		"ki_KE":          true,
		"kk":             true,
		"kk_Cyrl":        true,
		"kk_Cyrl_KZ":     true,
		"kkj":            true,
		"kkj_CM":         true,
		"kl":             true,
		"kl_GL":          true,
		"kln":            true,
		"kln_KE":         true,
		"km":             true,
		"km_KH":          true,
		"kn":             true,
		"kn_IN":          true,
		"ko":             true,
		"ko_KP":          true,
		"ko_KR":          true,
		"kok":            true,
		"kok_IN":         true,
		"ks":             true,
		"ks_Arab":        true,
		"ks_Arab_IN":     true,
		"ksb":            true,
		"ksb_TZ":         true,
		"ksf":            true,
		"ksf_CM":         true,
		"ksh":            true,
		"ksh_DE":         true,
		"kw":             true,
		"kw_GB":          true,
		"ky":             true,
		"ky_Cyrl":        true,
		"ky_Cyrl_KG":     true,
		"lag":            true,
		"lag_TZ":         true,
		"lg":             true,
		"lg_UG":          true,
		"lkt":            true,
		"lkt_US":         true,
		"ln":             true,
		"ln_AO":          true,
		"ln_CD":          true,
		"ln_CF":          true,
		"ln_CG":          true,
		"lo":             true,
		"lo_LA":          true,
		"lt":             true,
		"lt_LT":          true,
		"lu":             true,
		"lu_CD":          true,
		"luo":            true,
		"luo_KE":         true,
		"luy":            true,
		"luy_KE":         true,
		"lv":             true,
		"lv_LV":          true,
		"mas":            true,
		"mas_KE":         true,
		"mas_TZ":         true,
		"mer":            true,
		"mer_KE":         true,
		"mfe":            true,
		"mfe_MU":         true,
		"mg":             true,
		"mg_MG":          true,
		"mgh":            true,
		"mgh_MZ":         true,
		"mgo":            true,
		"mgo_CM":         true,
		"mk":             true,
		"mk_MK":          true,
		"ml":             true,
		"ml_IN":          true,
		"mn":             true,
		"mn_Cyrl":        true,
		"mn_Cyrl_MN":     true,
		"mr":             true,
		"mr_IN":          true,
		"ms":             true,
		"ms_Latn":        true,
		"ms_Latn_BN":     true,
		"ms_Latn_MY":     true,
		"ms_Latn_SG":     true,
		"mt":             true,
		"mt_MT":          true,
		"mua":            true,
		"mua_CM":         true,
		"my":             true,
		"my_MM":          true,
		"naq":            true,
		"naq_NA":         true,
		"nb":             true,
		"nb_NO":          true,
		"nb_SJ":          true,
		"nd":             true,
		"nd_ZW":          true,
		"ne":             true,
		"ne_IN":          true,
		"ne_NP":          true,
		"nl":             true,
		"nl_AW":          true,
		"nl_BE":          true,
		"nl_BQ":          true,
		"nl_CW":          true,
		"nl_NL":          true,
		"nl_SR":          true,
		"nl_SX":          true,
		"nmg":            true,
		"nmg_CM":         true,
		"nn":             true,
		"nn_NO":          true,
		"nnh":            true,
		"nnh_CM":         true,
		"nr":             true,
		"nr_ZA":          true,
		"nso":            true,
		"nso_ZA":         true,
		"nus":            true,
		"nus_SD":         true,
		"nyn":            true,
		"nyn_UG":         true,
		"om":             true,
		"om_ET":          true,
		"om_KE":          true,
		"or":             true,
		"or_IN":          true,
		"ordinals":       true,
		"os":             true,
		"os_GE":          true,
		"os_RU":          true,
		"pa":             true,
		"pa_Arab":        true,
		"pa_Arab_PK":     true,
		"pa_Guru":        true,
		"pa_Guru_IN":     true,
		"pl":             true,
		"pl_PL":          true,
		"plurals":        true,
		"ps":             true,
		"ps_AF":          true,
		"pt":             true,
		"pt_AO":          true,
		"pt_BR":          true,
		"pt_CV":          true,
		"pt_GW":          true,
		"pt_MO":          true,
		"pt_MZ":          true,
		"pt_PT":          true,
		"pt_ST":          true,
		"pt_TL":          true,
		"rm":             true,
		"rm_CH":          true,
		"rn":             true,
		"rn_BI":          true,
		"ro":             true,
		"ro_MD":          true,
		"ro_RO":          true,
		"rof":            true,
		"rof_TZ":         true,
		"ru":             true,
		"ru_BY":          true,
		"ru_KG":          true,
		"ru_KZ":          true,
		"ru_MD":          true,
		"ru_RU":          true,
		"ru_UA":          true,
		"rw":             true,
		"rw_RW":          true,
		"rwk":            true,
		"rwk_TZ":         true,
		"sah":            true,
		"sah_RU":         true,
		"saq":            true,
		"saq_KE":         true,
		"sbp":            true,
		"sbp_TZ":         true,
		"se":             true,
		"se_FI":          true,
		"se_NO":          true,
		"seh":            true,
		"seh_MZ":         true,
		"ses":            true,
		"ses_ML":         true,
		"sg":             true,
		"sg_CF":          true,
		"shi":            true,
		"shi_Latn":       true,
		"shi_Latn_MA":    true,
		"shi_Tfng":       true,
		"shi_Tfng_MA":    true,
		"si":             true,
		"si_LK":          true,
		"sk":             true,
		"sk_SK":          true,
		"sl":             true,
		"sl_SI":          true,
		"sn":             true,
		"sn_ZW":          true,
		"so":             true,
		"so_DJ":          true,
		"so_ET":          true,
		"so_KE":          true,
		"so_SO":          true,
		"sq":             true,
		"sq_AL":          true,
		"sq_MK":          true,
		"sq_XK":          true,
		"sr":             true,
		"sr_Cyrl":        true,
		"sr_Cyrl_BA":     true,
		"sr_Cyrl_ME":     true,
		"sr_Cyrl_RS":     true,
		"sr_Cyrl_XK":     true,
		"sr_Latn":        true,
		"sr_Latn_BA":     true,
		"sr_Latn_ME":     true,
		"sr_Latn_RS":     true,
		"sr_Latn_XK":     true,
		"ss":             true,
		"ss_SZ":          true,
		"ss_ZA":          true,
		"ssy":            true,
		"ssy_ER":         true,
		"st":             true,
		"st_LS":          true,
		"st_ZA":          true,
		"sv":             true,
		"sv_AX":          true,
		"sv_FI":          true,
		"sv_SE":          true,
		"sw":             true,
		"sw_KE":          true,
		"sw_TZ":          true,
		"sw_UG":          true,
		"swc":            true,
		"swc_CD":         true,
		"ta":             true,
		"ta_IN":          true,
		"ta_LK":          true,
		"ta_MY":          true,
		"ta_SG":          true,
		"te":             true,
		"te_IN":          true,
		"teo":            true,
		"teo_KE":         true,
		"teo_UG":         true,
		"tg":             true,
		"tg_Cyrl":        true,
		"tg_Cyrl_TJ":     true,
		"th":             true,
		"th_TH":          true,
		"ti":             true,
		"ti_ER":          true,
		"ti_ET":          true,
		"tig":            true,
		"tig_ER":         true,
		"tn":             true,
		"tn_BW":          true,
		"tn_ZA":          true,
		"to":             true,
		"to_TO":          true,
		"tr":             true,
		"tr_CY":          true,
		"tr_TR":          true,
		"ts":             true,
		"ts_ZA":          true,
		"twq":            true,
		"twq_NE":         true,
		"tzm":            true,
		"tzm_Latn":       true,
		"tzm_Latn_MA":    true,
		"ug":             true,
		"ug_Arab":        true,
		"ug_Arab_CN":     true,
		"uk":             true,
		"uk_UA":          true,
		"ur":             true,
		"ur_IN":          true,
		"ur_PK":          true,
		"uz":             true,
		"uz_Arab":        true,
		"uz_Arab_AF":     true,
		"uz_Cyrl":        true,
		"uz_Cyrl_UZ":     true,
		"uz_Latn":        true,
		"uz_Latn_UZ":     true,
		"vai":            true,
		"vai_Latn":       true,
		"vai_Latn_LR":    true,
		"vai_Vaii":       true,
		"vai_Vaii_LR":    true,
		"ve":             true,
		"ve_ZA":          true,
		"vi":             true,
		"vi_VN":          true,
		"vo":             true,
		"vo_001":         true,
		"vun":            true,
		"vun_TZ":         true,
		"wae":            true,
		"wae_CH":         true,
		"wal":            true,
		"wal_ET":         true,
		"xh":             true,
		"xh_ZA":          true,
		"xog":            true,
		"xog_UG":         true,
		"yav":            true,
		"yav_CM":         true,
		"yo":             true,
		"yo_BJ":          true,
		"yo_NG":          true,
		"zgh":            true,
		"zgh_MA":         true,
		"zh":             true,
		"zh_Hans":        true,
		"zh_Hans_CN":     true,
		"zh_Hans_HK":     true,
		"zh_Hans_MO":     true,
		"zh_Hans_SG":     true,
		"zh_Hant":        true,
		"zh_Hant_HK":     true,
		"zh_Hant_MO":     true,
		"zh_Hant_TW":     true,
		"zu":             true,
		"zu_ZA":          true,
	}

	// Class wide Locale Constants
	territoryData = map[string]string{
		"AD": "ca_AD",
		"AE": "ar_AE",
		"AF": "fa_AF",
		"AG": "en_AG",
		"AI": "en_AI",
		"AL": "sq_AL",
		"AM": "hy_AM",
		"AN": "pap_AN",
		"AO": "pt_AO",
		"AQ": "und_AQ",
		"AR": "es_AR",
		"AS": "sm_AS",
		"AT": "de_AT",
		"AU": "en_AU",
		"AW": "nl_AW",
		"AX": "sv_AX",
		"AZ": "az_Latn_AZ",
		"BA": "bs_BA",
		"BB": "en_BB",
		"BD": "bn_BD",
		"BE": "nl_BE",
		"BF": "mos_BF",
		"BG": "bg_BG",
		"BH": "ar_BH",
		"BI": "rn_BI",
		"BJ": "fr_BJ",
		"BL": "fr_BL",
		"BM": "en_BM",
		"BN": "ms_BN",
		"BO": "es_BO",
		"BR": "pt_BR",
		"BS": "en_BS",
		"BT": "dz_BT",
		"BV": "und_BV",
		"BW": "en_BW",
		"BY": "be_BY",
		"BZ": "en_BZ",
		"CA": "en_CA",
		"CC": "ms_CC",
		"CD": "sw_CD",
		"CF": "fr_CF",
		"CG": "fr_CG",
		"CH": "de_CH",
		"CI": "fr_CI",
		"CK": "en_CK",
		"CL": "es_CL",
		"CM": "fr_CM",
		"CN": "zh_Hans_CN",
		"CO": "es_CO",
		"CR": "es_CR",
		"CU": "es_CU",
		"CV": "kea_CV",
		"CX": "en_CX",
		"CY": "el_CY",
		"CZ": "cs_CZ",
		"DE": "de_DE",
		"DJ": "aa_DJ",
		"DK": "da_DK",
		"DM": "en_DM",
		"DO": "es_DO",
		"DZ": "ar_DZ",
		"EC": "es_EC",
		"EE": "et_EE",
		"EG": "ar_EG",
		"EH": "ar_EH",
		"ER": "ti_ER",
		"ES": "es_ES",
		"ET": "en_ET",
		"FI": "fi_FI",
		"FJ": "hi_FJ",
		"FK": "en_FK",
		"FM": "chk_FM",
		"FO": "fo_FO",
		"FR": "fr_FR",
		"GA": "fr_GA",
		"GB": "en_GB",
		"GD": "en_GD",
		"GE": "ka_GE",
		"GF": "fr_GF",
		"GG": "en_GG",
		"GH": "ak_GH",
		"GI": "en_GI",
		"GL": "iu_GL",
		"GM": "en_GM",
		"GN": "fr_GN",
		"GP": "fr_GP",
		"GQ": "fan_GQ",
		"GR": "el_GR",
		"GS": "und_GS",
		"GT": "es_GT",
		"GU": "en_GU",
		"GW": "pt_GW",
		"GY": "en_GY",
		"HK": "zh_Hant_HK",
		"HM": "und_HM",
		"HN": "es_HN",
		"HR": "hr_HR",
		"HT": "ht_HT",
		"HU": "hu_HU",
		"ID": "id_ID",
		"IE": "en_IE",
		"IL": "he_IL",
		"IM": "en_IM",
		"IN": "hi_IN",
		"IO": "und_IO",
		"IQ": "ar_IQ",
		"IR": "fa_IR",
		"IS": "is_IS",
		"IT": "it_IT",
		"JE": "en_JE",
		"JM": "en_JM",
		"JO": "ar_JO",
		"JP": "ja_JP",
		"KE": "en_KE",
		"KG": "ky_Cyrl_KG",
		"KH": "km_KH",
		"KI": "en_KI",
		"KM": "ar_KM",
		"KN": "en_KN",
		"KP": "ko_KP",
		"KR": "ko_KR",
		"KW": "ar_KW",
		"KY": "en_KY",
		"KZ": "ru_KZ",
		"LA": "lo_LA",
		"LB": "ar_LB",
		"LC": "en_LC",
		"LI": "de_LI",
		"LK": "si_LK",
		"LR": "en_LR",
		"LS": "st_LS",
		"LT": "lt_LT",
		"LU": "fr_LU",
		"LV": "lv_LV",
		"LY": "ar_LY",
		"MA": "ar_MA",
		"MC": "fr_MC",
		"MD": "ro_MD",
		"ME": "sr_Latn_ME",
		"MF": "fr_MF",
		"MG": "mg_MG",
		"MH": "mh_MH",
		"MK": "mk_MK",
		"ML": "bm_ML",
		"MM": "my_MM",
		"MN": "mn_Cyrl_MN",
		"MO": "zh_Hant_MO",
		"MP": "en_MP",
		"MQ": "fr_MQ",
		"MR": "ar_MR",
		"MS": "en_MS",
		"MT": "mt_MT",
		"MU": "mfe_MU",
		"MV": "dv_MV",
		"MW": "ny_MW",
		"MX": "es_MX",
		"MY": "ms_MY",
		"MZ": "pt_MZ",
		"NA": "kj_NA",
		"NC": "fr_NC",
		"NE": "ha_Latn_NE",
		"NF": "en_NF",
		"NG": "en_NG",
		"NI": "es_NI",
		"NL": "nl_NL",
		"NO": "nb_NO",
		"NP": "ne_NP",
		"NR": "en_NR",
		"NU": "niu_NU",
		"NZ": "en_NZ",
		"OM": "ar_OM",
		"PA": "es_PA",
		"PE": "es_PE",
		"PF": "fr_PF",
		"PG": "tpi_PG",
		"PH": "fil_PH",
		"PK": "ur_PK",
		"PL": "pl_PL",
		"PM": "fr_PM",
		"PN": "en_PN",
		"PR": "es_PR",
		"PS": "ar_PS",
		"PT": "pt_PT",
		"PW": "pau_PW",
		"PY": "gn_PY",
		"QA": "ar_QA",
		"RE": "fr_RE",
		"RO": "ro_RO",
		"RS": "sr_Cyrl_RS",
		"RU": "ru_RU",
		"RW": "rw_RW",
		"SA": "ar_SA",
		"SB": "en_SB",
		"SC": "crs_SC",
		"SD": "ar_SD",
		"SE": "sv_SE",
		"SG": "en_SG",
		"SH": "en_SH",
		"SI": "sl_SI",
		"SJ": "nb_SJ",
		"SK": "sk_SK",
		"SL": "kri_SL",
		"SM": "it_SM",
		"SN": "fr_SN",
		"SO": "sw_SO",
		"SR": "srn_SR",
		"ST": "pt_ST",
		"SV": "es_SV",
		"SY": "ar_SY",
		"SZ": "en_SZ",
		"TC": "en_TC",
		"TD": "fr_TD",
		"TF": "und_TF",
		"TG": "fr_TG",
		"TH": "th_TH",
		"TJ": "tg_Cyrl_TJ",
		"TK": "tkl_TK",
		"TL": "pt_TL",
		"TM": "tk_TM",
		"TN": "ar_TN",
		"TO": "to_TO",
		"TR": "tr_TR",
		"TT": "en_TT",
		"TV": "tvl_TV",
		"TW": "zh_Hant_TW",
		"TZ": "sw_TZ",
		"UA": "uk_UA",
		"UG": "sw_UG",
		"UM": "en_UM",
		"US": "en_US",
		"UY": "es_UY",
		"UZ": "uz_Cyrl_UZ",
		"VA": "it_VA",
		"VC": "en_VC",
		"VE": "es_VE",
		"VG": "en_VG",
		"VI": "en_VI",
		"VN": "vi_VN",
		"VU": "bi_VU",
		"WF": "wls_WF",
		"WS": "sm_WS",
		"YE": "ar_YE",
		"YT": "swb_YT",
		"ZA": "en_ZA",
		"ZM": "en_ZM",
		"ZW": "sn_ZW",
	}
)

// Locale is a locale
type Locale struct {
	locale      string
	auto        map[string]string
	browser     map[string]string
	environment map[string]string
	breakChain  bool
	def         map[string]string
}

// SetLocale sets new locale
func (l *Locale) SetLocale(locale string) error {
	locale, err := l.prepareLocale(locale, false)
	if err != nil {
		return err
	}

	if !utils.MapSBKeyExists(locale, localeData) {
		// Is it an alias? If so, we can use this locale
		if utils.MapSSKeyExists(locale, localeAliases) {
			l.locale = locale
			return nil
		}

		region := ""
		if len(locale) > 2 {
			region = locale[0:3]
			if region[2:3] == "_" || region[2:3] == "-" {
				region = region[0:2]
			}
		}

		if utils.MapSBKeyExists(region, localeData) {
			l.locale = region
		} else {
			l.locale = "root"
		}
	} else {
		l.locale = locale
	}

	return nil
}

// Locale returns locale string
func (l *Locale) Locale() string {
	return l.locale
}

// Language returns the language part of the locale
func (l *Locale) Language() string {
	locale := strings.Split(l.locale, "_")
	return locale[0]
}

// Region returns the region part of the locale if available
func (l *Locale) Region() string {
	locale := strings.Split(l.locale, "_")
	if len(locale) > 1 {
		return locale[1]
	}

	return ""
}

func (l *Locale) prepareLocale(locale string, strict bool) (string, error) {
	if len(l.auto) == 0 {
		l.browser, _ = Browser()
		l.environment, _ = Environment()
		l.breakChain = true
		l.auto = utils.MapSSMerge(l.browser, l.environment)
	}

	var localeData map[string]string
	if !strict {
		if locale == "browser" {
			localeData = l.browser
		} else if locale == "environment" {
			localeData = l.environment
		} else if locale == "default" {
			localeData = l.def
		} else if locale == "auto" || locale == "" {
			localeData = l.auto
		}

		for _, v := range localeData {
			locale = v
			break
		}
	}

	// This can only happen when someone extends Zend_Locale and erases the default
	if locale == "" {
		return "", errors.New("Autodetection of Locale has been failed")
	}

	if i := strings.Index(locale, "-"); i > -1 {
		locale = strings.Replace(locale, "-", "_", -1)
	}

	parts := strings.Split(locale, "_")
	if _, ok := localeData[parts[0]]; !ok {
		if len(parts) == 1 && utils.MapSSKeyExists(parts[0], territoryData) {
			return territoryData[parts[0]], nil
		}

		return "", nil
	}

	for key, value := range parts {
		if len(value) < 2 || len(value) > 3 {
			parts = append(parts[:key], parts[key:]...)
		}
	}

	locale = strings.Join(parts, "_")
	return locale, nil
}

// NewLocale create a new locale
func NewLocale(locale string) (*Locale, error) {
	l := &Locale{}
	if err := l.SetLocale(locale); err != nil {
		return nil, errors.Wrap(err, "Unable to create locale")
	}

	return l, nil
}

// Browser return a map of all accepted languages of the client
// Expects RFC compilant Header !!
//
// The notation can be :
// de,en-UK-US;q=0.5,fr-FR;q=0.2
func Browser() (map[string]string, error) {
	/*$httplanguages = getenv('HTTP_ACCEPT_LANGUAGE');
	if (empty($httplanguages) && array_key_exists('HTTP_ACCEPT_LANGUAGE', $_SERVER)) {
		$httplanguages = $_SERVER['HTTP_ACCEPT_LANGUAGE'];
	}

	$languages     = array();
	if (empty($httplanguages)) {
		return $languages;
	}*/

	//$accepted = preg_split('/,\s*/', $httplanguages);

	/*foreach ($accepted as $accept) {
		$match  = null;
		$result = preg_match('/^([a-z]{1,8}(?:[-_][a-z]{1,8})*)(?:;\s*q=(0(?:\.[0-9]{1,3})?|1(?:\.0{1,3})?))?$/i',
								$accept, $match);

		if ($result < 1) {
			continue;
		}

		if (isset($match[2]) === true) {
			$quality = (float) $match[2];
		} else {
			$quality = 1.0;
		}

		$countrys = explode('-', $match[1]);
		$region   = array_shift($countrys);

		$country2 = explode('_', $region);
		$region   = array_shift($country2);

		foreach ($countrys as $country) {
			$languages[$region . '_' . strtoupper($country)] = $quality;
		}

		foreach ($country2 as $country) {
			$languages[$region . '_' . strtoupper($country)] = $quality;
		}

		if ((isset($languages[$region]) === false) || ($languages[$region] < $quality)) {
			$languages[$region] = $quality;
		}
	}

	self::$_browser = $languages;
	return $languages;*/
	return make(map[string]string), nil
}

// Environment expects the Systems standard locale
// For Windows:
// f.e.: LC_COLLATE=C;LC_CTYPE=German_Austria.1252;LC_MONETARY=C
// would be recognised as de_AT
func Environment() (map[string]string, error) {
	/*$language      = setlocale(LC_ALL, 0);
	$languages     = explode(';', $language);
	$languagearray = array();

	foreach ($languages as $locale) {
		if (strpos($locale, '=') !== false) {
			$language = substr($locale, strpos($locale, '='));
			$language = substr($language, 1);
		}

		if ($language !== 'C') {
			if (strpos($language, '.') !== false) {
				$language = substr($language, 0, strpos($language, '.'));
			} else if (strpos($language, '@') !== false) {
				$language = substr($language, 0, strpos($language, '@'));
			}

			$language = str_ireplace(
				array_keys(Zend_Locale_Data_Translation::$languageTranslation),
				array_values(Zend_Locale_Data_Translation::$languageTranslation),
				(string) $language
			);

			$language = str_ireplace(
				array_keys(Zend_Locale_Data_Translation::$regionTranslation),
				array_values(Zend_Locale_Data_Translation::$regionTranslation),
				$language
			);

			if (isset(self::$_localeData[$language]) === true) {
				$languagearray[$language] = 1;
				if (strpos($language, '_') !== false) {
					$languagearray[substr($language, 0, strpos($language, '_'))] = 1;
				}
			}
		}
	}

	self::$_environment = $languagearray;
	return $languagearray;*/
	return make(map[string]string), nil
}

// FindLocale finds the proper locale based on the input
func FindLocale(locale string) (string, error) {
	if locale == "" {
		var l *Locale
		lcl := registry.Get("WSFLocale")
		if lcl == nil {
			var err error
			l, err = NewLocale("")
			if err != nil {
				return "", err
			}
		} else {
			l = lcl.(*Locale)
		}

		locale = l.Locale()
	}

	if IsLocale(locale, true) {
		if !IsLocale(locale, false) {
			locale = ToTerritory(locale)

			if locale == "" {
				return "", errors.Errorf("The locale '%s' is no known locale", locale)
			}
		} else {
			lcl, err := NewLocale(locale)
			if err != nil {
				return "", err
			}
			locale = lcl.Locale()
		}
	}

	return prepareLocale(locale, false)
}

// ToTerritory returns teritorial locale
func ToTerritory(territory string) string {
	territory = strings.ToUpper(territory)
	if v, ok := territoryData[territory]; ok {
		return v
	}

	return ""
}

// IsLocale returns true if loc is a valid locale
func IsLocale(loc string, strict bool) bool {
	if loc == "" {
		return false
	}

	// Is it an alias?
	if _, ok := localeAliases[loc]; ok {
		return true
	}

	var err error
	loc, err = prepareLocale(loc, strict)
	if err != nil {
		return false
	}

	if _, ok := localeData[loc]; ok {
		return true
	} else if !strict {
		parts := strings.Split(loc, "_")
		if _, ok := localeData[parts[0]]; ok {
			return true
		}
	}

	return false
}

func prepareLocale(locale string, strict bool) (string, error) {
	/*if len(auto) == 0 {
		browser     = Browser()
		environment = Environment()
		breakChain  = true
		auto        = self::getBrowser() + self::getEnvironment() + self::getDefault();
	}

	lcl := make(map[string]string)
	if !strict {
		if locale == "browser" {
			lcl = browser
		}

		if locale == "environment" {
			lcl = environment
		}

		if locale == "default" {
			lcl = def
		}

		if locale == "auto" || locale == "" {
			lcl = auto
		}
	}

	if (strpos($locale, '-') !== false) {
		$locale = strtr($locale, '-', '_');
	}

	$parts = explode('_', $locale);
	if (!isset(self::$_localeData[$parts[0]])) {
		if ((count($parts) === 1) && array_key_exists($parts[0], self::$_territoryData)) {
			return self::$_territoryData[$parts[0]];
		}

		return '';
	}

	foreach($parts as $key => $value) {
		if ((strlen($value) < 2) || (strlen($value) > 3)) {
			unset($parts[$key]);
		}
	}

	$locale = implode('_', $parts);
	return (string) $locale;*/
	return locale, nil
}
