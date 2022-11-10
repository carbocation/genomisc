# This software is released under the following BSD 3-Clause License:

# Copyright (c) 2022 James Pirruccello and The General Hospital Corporation.
# All rights reserved.

# Redistribution and use in source and binary forms, with or without
# modification, are permitted provided that the following conditions are met:

# 1. Redistributions of source code must retain the above copyright notice, this
#    list of conditions and the following disclaimer.

# 2. Redistributions in binary form must reproduce the above copyright notice,
#    this list of conditions and the following disclaimer in the documentation
#    and/or other materials provided with the distribution.

# 3. Neither the name of the copyright holder nor the names of its contributors
#    may be used to endorse or promote products derived from this software
#    without specific prior written permission.

# THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
# AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
# IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
# DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
# FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
# DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
# SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
# CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
# OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
# OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

# AORTA Score. This file represents the R program for computing the AORTA Score
# v20220503. The script contains the model coefficients that were derived by
# James Pirruccello against 30,018 UK Biobank aortic measurements using the
# glinternet package (glinternet is by Lim and Hastie 2015). The manuscript
# describing the derivation of the AORTA Score can be found at
# doi:10.1001/jama.2022.19701 .

# The main function is `aorta.score`, which expects a dataframe. E.g.:
#
# aorta.score(dat)

# To apply the function to a dataframe of people, and save the predicted aortic
# diameter as a column:
#
# dat$predicted_ascending_aorta_diam <- aorta.score(dat)

# The variables in your dataframe need to have the same names as those that were
# used in the model. 

# Missing variables are reported as a warning. So to learn the expected column
# names, one can invoke:
# 
# aorta.score(0)
# 
# ... and the warning will list the column names
#
# Also note that although BMI is fully determined by height and weight, the
# score expects you to have computed BMI and does not check that your BMI
# calculation was correct.
#
# Finally, please note that this software does not constrain either the inputs
# or the outputs to be within the training set range or even within a
# physiologic range. As we note on aortascore.com: This tool is not a medical
# device, nor does it provide a medical or diagnostic service. The results are
# intended for use by healthcare professionals for informational and educational
# purposes only. The information provided is not an attempt to practice medicine
# or provide specific medical advice, and should not be used to make a diagnosis
# or in the place of a qualified healthcare professional's clinical judgment.
# You assume full responsibility for using the information provided by the tool,
# and you understand and agree that the Broad Institute, Inc. and its affiliates
# and subsidiaries are not responsible or liable for any claim, loss, damage or
# injury, including but not limited to death, resulting from the use of this
# tool by you or any user. We do not guarantee the validity, completeness,
# accuracy or timeliness of information provided by the tool or that access to
# the website will be error- or virus-free. We disclaim any warranty, whether
# express or implied, including warranties of merchantability or fitness for a
# particular purpose.

aorta.score.version <- "v20220503"

aorta.score <- function(dat) {
    categorical.categorical <- read.table(text = "level1	level2	value	namedCatEff1	namedCatEff2
0	0	0.00186363994907011	sex_Female	prevalent_dm
1	0	-0.00186363994907011	sex_Female	prevalent_dm
0	1	-0.00186363994907011	sex_Female	prevalent_dm
1	1	0.00186363994907011	sex_Female	prevalent_dm
0	0	-0.0135521189316198	prevalent_dm	prevalent_htn
1	0	0.0135521189316198	prevalent_dm	prevalent_htn
0	1	0.0135521189316198	prevalent_dm	prevalent_htn
1	1	-0.0135521189316199	prevalent_dm	prevalent_htn
0	0	0.0021412691409754	prevalent_dm	prevalent_hld
1	0	-0.00214126914097541	prevalent_dm	prevalent_hld
0	1	-0.00214126914097541	prevalent_dm	prevalent_hld
1	1	0.0021412691409754	prevalent_dm	prevalent_hld
", header = TRUE, stringsAsFactors = FALSE)

    categorical.continuous <- read.table(text = "value	level	namedCatEff	namedContEff
0.000709345388250865	0	sex_Female	age
-0.000709345388250865	1	sex_Female	age
0.000415477270701555	0	prevalent_htn	age
-0.000415477270701555	1	prevalent_htn	age
0.00426227015499906	0	sex_Female	bmi
-0.00426227015499906	1	sex_Female	bmi
0.000461003686316868	0	prevalent_htn	bmi
-0.000461003686316868	1	prevalent_htn	bmi
-3.43653541027004e-05	0	prevalent_htn	pulse_rate
3.43653541027004e-05	1	prevalent_htn	pulse_rate
-2.55109727239257e-05	0	prevalent_hld	pulse_rate
2.55109727239257e-05	1	prevalent_hld	pulse_rate
0.000193132350452402	0	prevalent_htn	sbp
-0.000193132350452402	1	prevalent_htn	sbp
2.78943389341104e-05	0	prevalent_hld	sbp
-2.78943389341104e-05	1	prevalent_hld	sbp
-0.00025380347323269	0	sex_Female	dbp
0.00025380347323269	1	sex_Female	dbp
-0.000130563874958604	0	sex_Female	height_cm
0.000130563874958604	1	sex_Female	height_cm
4.41307037835025e-05	0	prevalent_htn	weight_kg
-4.41307037835025e-05	1	prevalent_htn	weight_kg
", header = TRUE, stringsAsFactors = FALSE)

    categorical.main <- read.table(text = "value	level	namedCatEff
-0.20823065887864	0	sex_Female
-0.0499987689936839	1	sex_Female
0.0117256871576651	0	prevalent_dm
-0.011725687157665	1	prevalent_dm
-0.0952421519461697	0	prevalent_htn
0.0854636151258379	1	prevalent_htn
0.00543187576739287	0	prevalent_hld
0.00396040385066557	1	prevalent_hld
", header = TRUE, stringsAsFactors = FALSE)

    continuous.continuous <- read.table(text = "value	namedContEff1	namedContEff2
-6.74790866190736e-05	age	sbp
3.59345287796709e-05	age	dbp
2.04636585480731e-05	bmi	pulse_rate
-1.26979072744856e-05	bmi	sbp
-5.51223322739197e-05	bmi	weight_kg
-2.16393282872501e-05	pulse_rate	sbp
-2.85004004331634e-07	pulse_rate	height_cm
-8.05918464737499e-07	sbp	dbp
-5.58212143854082e-06	sbp	weight_kg
-4.47124731552925e-06	dbp	weight_kg
", header = TRUE, stringsAsFactors = FALSE)

    continuous.main <- read.table(text = "value	effect	namedContEff
0.0169283018283151	1	age
-0.000543460207844996	3	pulse_rate
0.00494314572832139	4	sbp
0.00632391449981155	5	dbp
0.00691587091058802	6	height_cm
0.00611489831591748	7	weight_kg
0.00690412743686394	2	bmi
", header = TRUE, stringsAsFactors = FALSE)

    intercept <- read.table(text = "Intercept
0.107160128479571
", header = TRUE, stringsAsFactors = FALSE)

    required.columns <- function() {
        unique(c(
            continuous.main$namedContEff,
            continuous.continuous$namedContEff1,
            continuous.continuous$namedContEff2,
            categorical.main$namedCatEff,
            categorical.continuous$namedCatEff,
            categorical.continuous$namedContEff,
            categorical.categorical$namedCatEff1,
            categorical.categorical$namedCatEff2
        ))
    }
    
    # Make sure that the data frame has all columns present
    have.names <- names(dat)
    missing.names <- required.columns()[!(required.columns() %in% have.names)]
    if(length(missing.names) > 0){
        warning("Missing columns: {", paste0(missing.names, collapse = ", "), "}")
        stop(".")
    }
    
    # intercept
    score <- rep(0, nrow(dat)) + intercept$Intercept
    
    # categorical.main
    for(i in 1:nrow(categorical.main)) {
        row <- categorical.main[i,]

        score <- score + (dat[[row$namedCatEff]] == row$level) * row$value
    }
    
    # continuous.main
    for(i in 1:nrow(continuous.main)) {
        row <- continuous.main[i,]

        score <- score + dat[[row$namedContEff]] * row$value
    }
    
    # categorical.categorical
    for(i in 1:nrow(categorical.categorical)) {
        row <- categorical.categorical[i,]

        score <- score + (dat[[row$namedCatEff1]] == row$level1) * (dat[[row$namedCatEff2]] == row$level2) * row$value
    }
    
    # continuous.continuous
    for(i in 1:nrow(continuous.continuous)) {
        row <- continuous.continuous[i,]

        score <- score + (dat[[row$namedContEff1]]) * (dat[[row$namedContEff2]]) * row$value
    }
    
    # categorical.continuous
    for(i in 1:nrow(categorical.continuous)) {
        row <- categorical.continuous[i,]

        score <- score + (dat[[row$namedCatEff]] == row$level) * (dat[[row$namedContEff]]) * row$value
    }
    
    score
}
