// intersect_arm.s
// vim: ft=9asm ai

// func intersect(a, b, c, rDir, rFrom vec64.Vector) (rayDepth, u, v float64)
TEXT ·intersect(SB),7,$0
    B           ·intersect_go(SB)
