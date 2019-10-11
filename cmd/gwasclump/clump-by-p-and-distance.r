list.of.packages <- c("ggplot2", "Hmisc", "data.table", "speedglm", "data.table", "qqman", "CMplot", "optparse")
new.packages <- list.of.packages[!(list.of.packages %in% installed.packages()[,"Package"])]
if(length(new.packages)) install.packages(new.packages,repos = "http://cran.us.r-project.org")

library("optparse")
library("speedglm")
library("ggplot2")
library("data.table")
library("qqman")
library("CMplot")

option_list = list(
    make_option(c("-f", "--file"), type="character", default=NULL, help="BOLT output file"),
    make_option("--pvalue", type="numeric", default=5e-8, help="P value threshold, below which loci are considered significant"),
    make_option(c("-r", "--locusRadius"), type="numeric", default=NULL, help="Number of bases around each lead SNP to exclude other potential lead SNPs")
)

opt <- parse_args(OptionParser(option_list=option_list))

print(opt$file)

stop("OK")

snpDistanceCutoff <- function(pos, bases) {
    lower = pos-bases
    if(lower<1){
        lower = 1
    }
    
    upper = pos+bases
    
    list(lower = lower, upper = upper)
}

pruneSNPs <- function(dat, flanking = 5e5, significance = 5e-8) {
    # Exit early if none are left in the input
    if(nrow(dat)==0){
        return()
    }
    
    # The best remaining SNP
    best <- subset(dat, is.finite(CHISQ_BOLT_LMM) & (CHISQ_BOLT_LMM == max(dat$CHISQ_BOLT_LMM)))
    if(nrow(best)==0){
        return()
    }
    
    # If the best remaining SNP is outside of your permitted bounds, exit
    if(best$P_BOLT_LMM > significance) {
        return()
    }
    
    # Identify the window around the SNP
    dist = snpDistanceCutoff(best$BP, flanking)
    
    # Return the SNP and the remaining data with the 
    # flanking SNPs around the newest best SNP removed
    list(best = best, remaining = subset(dat, CHR != best$CHR | BP < dist$lower | BP > dist$upper ))
}

iteratively.prune <- function(df, flanking=1e6, significance=1e-7) {
    bestSNPs <- list()

    testDat <- df
    i <- 0
    repeat {
        i <- i + 1
        res <- pruneSNPs(testDat, flanking, significance)
        if(is.null(res)) {
            break
        }

        bestSNPs[[i]] <- res$best
        testDat <- res$remaining
    }

    all.best <- do.call(rbind, bestSNPs)
    if(is.data.frame(all.best) && nrow(all.best) > 0) {
        all.best <- all.best[with(all.best, order(CHR, BP)),]
    }
    all.best
}

P <- "5e-7"
N <- "29041"

input.file <- sprintf(
    "gsutil cat gs://ukbb_v2/projects/jamesp/projects/gwas/lv20k/results/v42/corrected_extracted_%%s.bgen.stats.gz | gunzip -c | awk 'NR==1; $12>0 && $15>0 && $16<%s && ( ($7<0.5 && $8*$7*2*%s > 100) || ($7>0.5 && $8*(1.0-$7)*2*%s > 100) ) {print $0}'", 
    P,
    N,
    N)

input.file

traits <- c("lvef", "lvedv", "lvesv", "sv", "lvedvi", "lvesvi", "svi")
# traits <- c("svi")

res <- list()
i <- 1
for(trait in traits) {
    df <- fread(cmd=sprintf(input.file, trait))
    
    for(cutoff in c(5e-7, 5e-8, 5e-9, 5e-10)){
        i <- i+1
        
        N_Loci = nrow(iteratively.prune(df, significance = cutoff))
        if(is.null(N_Loci)) {
            N_Loci = 0
        } else if(!is.finite(N_Loci)) {
            N_Loci = 0
        }

        res[[i]] <- data.frame(
            Trait = trait,
            Cutoff = cutoff,
            N_Loci = N_Loci
        )
    }
}

do.call("rbind", res)

results <- do.call("rbind", res)

wide.results <- reshape(results, idvar = "Trait", timevar = "Cutoff", direction = "wide")
wide.results

addmargins(as.matrix(wide.results[,c(2,3,4,5)]))
