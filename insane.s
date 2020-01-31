#include "go_asm.h"
#include "textflag.h"

TEXT ·InsaneSkipWC(SB), NOSPLIT, $0-40
	MOVQ b_base+0(FP), SI
	MOVQ b_len+8(FP), BX
	MOVB c+24(FP), AL
	LEAQ ret+32(FP), R8

    MOVD AX, X0
    PUNPCKLBW X0, X0
    PUNPCKLBW X0, X0
    PSHUFL $0, X0, X0

    MOVQ SI, DI

	MOVD AX, X0
	LEAQ -32(SI)(BX*1), R11
	VPBROADCASTB  X0, Y1

avx2_loop:
	VMOVDQU (DI), Y2
	VPCMPEQB Y1, Y2, Y3
	VPTEST Y3, Y3
	JNZ avx2success
	ADDQ $32, DI
	CMPQ DI, R11
	JLT avx2_loop
	MOVQ R11, DI
	VMOVDQU (DI), Y2
	VPCMPEQB Y1, Y2, Y3
	VPTEST Y3, Y3
	JNZ avx2success
	VZEROUPPER
	MOVQ $-101, (R8)
	RET

avx2success:
	VPMOVMSKB Y3, DX
	BSFL DX, DX
	SUBQ SI, DI
	ADDQ DI, DX
	MOVQ DX, (R8)
	VZEROUPPER
	RET


TEXT ·InsaneSkipWC_(SB), NOSPLIT, $0-56
    // first slice is 0-23 byte
	MOVQ a_base+0(FP), SI
	MOVQ a_len+8(FP), AX

    // second slice is 24-47 byte
	MOVQ b_base+24(FP), DI
	MOVQ b_len+32(FP), BX

    // ret is after two slices with 48 byte offset
	LEAQ ret_a+48(FP), R8
	LEAQ ret_b+56(FP), R9

	VMOVDQU (DI), Y1

    MOVQ SI, DI
	LEAQ -32(SI)(BX*1), R11

loop:
	VMOVDQU (DI), Y2
	VPCMPEQB Y1, Y2, Y3
	VPTEST Y3, Y3
	JNZ success
	ADDQ $32, DI
	CMPQ DI, R11
	JLT loop
	MOVQ R11, DI
	VMOVDQU (DI), Y2
	VPCMPEQB Y1, Y2, Y3
	VPTEST Y3, Y3
	JNZ success
	VZEROUPPER
	MOVQ $-101, (R8)
	RET

success:
	VPMOVMSKB Y3, DX
	BSFL DX, DX
	SUBQ SI, DI
	ADDQ DI, DX
	MOVQ DX, (R8)
	MOVQ DX, (R9)
	VZEROUPPER
	RET

