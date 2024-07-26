# This software is released under the following BSD 3-Clause License:

# Copyright (c) 2023 James Pirruccello and The General Hospital Corporation.
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

# The source publication for this code is 10.1093/eurheartj/ehae474 .

# The main function is `aorta.gene`, which expects a dataframe. E.g.:
#
# aorta.gene(dat)

# To apply the function to a dataframe of people, and save the predicted aortic
# diameter as a column:
#
# dat$predicted_ascending_aorta_diam <- aorta.gene(dat)

# The variables in your dataframe need to have the same names as those that were
# used in the model. Note that "globally_residualized_score" is the polygenic
# score result after residualizing for PCs in your cohort (or a simliar cohort).

# Missing variables are reported as a warning. So to learn the expected column
# names, one can invoke:
# 
# aorta.gene(0)
# 
# ... and the warning will list the column names
#
# Also note that although BMI is fully determined by height and weight, the
# score expects you to have computed BMI and does not check that your BMI
# calculation was correct.
#
# Finally, please note that this software does not constrain either the inputs
# or the outputs to be within the training set range or even within a
# physiologic range. This tool is not a medical device, nor does it provide a
# medical or diagnostic service. The results are intended for use by healthcare
# professionals for informational and educational purposes only. The information
# provided is not an attempt to practice medicine or provide specific medical
# advice, and should not be used to make a diagnosis or in the place of a
# qualified healthcare professional's clinical judgment. You assume full
# responsibility for using the information provided by the tool, and you
# understand and agree that the Broad Institute, Inc. and its affiliates and
# subsidiaries are not responsible or liable for any claim, loss, damage or
# injury, including but not limited to death, resulting from the use of this
# tool by you or any user. We do not guarantee the validity, completeness,
# accuracy or timeliness of information provided by the tool or that access to
# the website will be error- or virus-free. We disclaim any warranty, whether
# express or implied, including warranties of merchantability or fitness for a
# particular purpose.

aorta.gene.version <- "v20230529_imputed_v2"

aorta.gene <- function(dat) {
    categorical.categorical <- read.table(text = "level1	level2	value	namedCatEff1	namedCatEff2
0	0	0.00354000206268387	sex_Female	prevalent_dm
1	0	-0.00354000206268395	sex_Female	prevalent_dm
0	1	-0.00354000206268395	sex_Female	prevalent_dm
1	1	0.00354000206268388	sex_Female	prevalent_dm
0	0	-0.000494365295753715	sex_Female	prevalent_hld
1	0	0.000494365295753698	sex_Female	prevalent_hld
0	1	0.000494365295753698	sex_Female	prevalent_hld
1	1	-0.000494365295753715	sex_Female	prevalent_hld
0	0	-0.00750489415333711	prevalent_dm	prevalent_htn
1	0	0.0075048941533364	prevalent_dm	prevalent_htn
0	1	0.0075048941533364	prevalent_dm	prevalent_htn
1	1	-0.00750489415333712	prevalent_dm	prevalent_htn
0	0	0.00374346164254748	prevalent_dm	prevalent_hld
1	0	-0.0037434616425476	prevalent_dm	prevalent_hld
0	1	-0.0037434616425476	prevalent_dm	prevalent_hld
1	1	0.00374346164254748	prevalent_dm	prevalent_hld
0	0	0.000458736770891237	prevalent_htn	prevalent_hld
1	0	-0.000458736770891233	prevalent_htn	prevalent_hld
0	1	-0.000458736770891233	prevalent_htn	prevalent_hld
1	1	0.000458736770891237	prevalent_htn	prevalent_hld
", header = TRUE, stringsAsFactors = FALSE)

    categorical.continuous <- read.table(text = "value	level	namedCatEff	namedContEff
0.00106057871213138	0	sex_Female	age
-0.00106057871213138	1	sex_Female	age
0.000177308576546309	0	prevalent_dm	age
-0.000177308576546309	1	prevalent_dm	age
-0.000816701834976061	0	prevalent_hld	age
0.000816701834976061	1	prevalent_hld	age
1.84537735833713e-06	0	prevalent_htn	age2
-1.84537735833713e-06	1	prevalent_htn	age2
0.00471120766065946	0	sex_Female	bmi
-0.00471120766065946	1	sex_Female	bmi
-0.000150833183424846	0	prevalent_dm	bmi
0.000150833183424846	1	prevalent_dm	bmi
0.00018069285774045	0	prevalent_htn	bmi
-0.00018069285774045	1	prevalent_htn	bmi
-0.000772332644212867	0	prevalent_hld	bmi
0.000772332644212867	1	prevalent_hld	bmi
-7.32896684106607e-05	0	sex_Female	pulse_rate
7.32896684106607e-05	1	sex_Female	pulse_rate
-0.000196400720867547	0	prevalent_htn	pulse_rate
0.000196400720867547	1	prevalent_htn	pulse_rate
-7.55814952030389e-05	0	prevalent_hld	pulse_rate
7.55814952030389e-05	1	prevalent_hld	pulse_rate
5.35346131716757e-05	0	sex_Female	sbp
-5.35346131716757e-05	1	sex_Female	sbp
-7.74831681449223e-05	0	prevalent_dm	sbp
7.74831681449223e-05	1	prevalent_dm	sbp
0.000622264490386452	0	prevalent_htn	sbp
-0.000622264490386452	1	prevalent_htn	sbp
-0.000310804386114275	0	sex_Female	dbp
0.000310804386114275	1	sex_Female	dbp
-0.000740108874188454	0	prevalent_htn	dbp
0.000740108874188454	1	prevalent_htn	dbp
0.000124457973640786	0	prevalent_hld	dbp
-0.000124457973640786	1	prevalent_hld	dbp
-0.000137237361355264	0	sex_Female	height_cm
0.000137237361355264	1	sex_Female	height_cm
0.000390012687756373	0	prevalent_htn	weight_kg
-0.000390012687756373	1	prevalent_htn	weight_kg
-7.87131942327662e-05	0	prevalent_hld	weight_kg
7.87131942327662e-05	1	prevalent_hld	weight_kg
", header = TRUE, stringsAsFactors = FALSE)

    categorical.main <- read.table(text = "value	level	namedCatEff
-0.233480489338584	0	sex_Female
-0.00654409578459022	1	sex_Female
0.0228149249249973	0	prevalent_dm
-0.0260225052293294	1	prevalent_dm
0.0377568045592725	0	prevalent_hld
-0.130565061931067	1	prevalent_hld
-0.0889927501153747	0	prevalent_htn
0.0938192085782183	1	prevalent_htn
", header = TRUE, stringsAsFactors = FALSE)

    continuous.continuous <- read.table(text = "value	namedContEff1	namedContEff2
-1.26505913619396e-06	age	age2
8.67377300140611e-05	age	dbp
-7.3387420355846e-07	age2	bmi
-6.01302263740941e-07	age2	sbp
-1.95161346945546e-07	age2	weight_kg
-2.53563671065889e-06	bmi	sbp
-1.91649248922501e-05	bmi	dbp
1.24204213927021e-05	bmi	height_cm
-8.63294017754071e-05	bmi	weight_kg
-2.53331946609657e-05	pulse_rate	sbp
1.57661740979341e-05	pulse_rate	weight_kg
4.3436145947803e-06	sbp	dbp
-1.17961882049952e-05	sbp	weight_kg
1.28081082926723e-05	dbp	height_cm
-1.08251885245117e-05	dbp	weight_kg
", header = TRUE, stringsAsFactors = FALSE)

    continuous.main <- read.table(text = "value	effect	namedContEff
0.000260271960029516	2	age2
-0.000838093204882467	4	pulse_rate
0.00310281704179048	5	sbp
0.00134078546828941	6	dbp
0.00584896555792012	7	height_cm
0.00799942617701697	8	weight_kg
0.00229829328598681	1	age
0.0122273032802419	3	bmi
", header = TRUE, stringsAsFactors = FALSE)

    intercept <- read.table(text = "Intercept
0.406267819015139
", header = TRUE, stringsAsFactors = FALSE)

    required.columns <- function() {
        unique(c(
            'globally_residualized_score',
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
    for(i in seq_len(nrow(categorical.main))) {
        row <- categorical.main[i,]

        score <- score + (dat[[row$namedCatEff]] == row$level) * row$value
    }
    
    # continuous.main
    for(i in seq_len(nrow(continuous.main))) {
        row <- continuous.main[i,]

        score <- score + dat[[row$namedContEff]] * row$value
    }
    
    # categorical.categorical
    for(i in seq_len(nrow(categorical.categorical))) {
        row <- categorical.categorical[i,]

        score <- score + (dat[[row$namedCatEff1]] == row$level1) * (dat[[row$namedCatEff2]] == row$level2) * row$value
    }
    
    # continuous.continuous
    for(i in seq_len(nrow(continuous.continuous))) {
        row <- continuous.continuous[i,]

        score <- score + (dat[[row$namedContEff1]]) * (dat[[row$namedContEff2]]) * row$value
    }
    
    # categorical.continuous
    for(i in seq_len(nrow(categorical.continuous))) {
        row <- categorical.continuous[i,]

        score <- score + (dat[[row$namedCatEff]] == row$level) * (dat[[row$namedContEff]]) * row$value
    }

    # Linear model incorporating the clinical components (`score`) and genetic
    # component (`dat$globally_residualized_score`):
    score <- 0.066076 + 0.979904 * score + 0.111250 * dat$globally_residualized_score

    score
}
