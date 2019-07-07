-- Demo query to reduce all scores so that each person has one score per trait
SELECT s.sample_id, source, SUM(score) score, SUM(n_incremented) n_incremented
FROM `ukbb-analyses.jpp_201907.v42_prs` prs
JOIN `ukbb-analyses.ukbb7089_201904.sample` s ON s.file_row = prs.sample_file_row 
GROUP BY s.sample_id, source
ORDER BY source ASC, s.sample_id ASC