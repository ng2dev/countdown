package main

import (
	"testing"

	"golang.org/x/crypto/ed25519"
)

func TestKeygen(t *testing.T) {
	const mnemonic = `shy else mystery outer define there front bracket dawn honey excuse virus lazy book kiss cannon oven law coconut hedgehog veteran narrow great cage`

	cases := map[string]string{
		"m/44'/988'/0'": "COUNTDOWN1h7rpratsyt7mylq79pakjfdzg839zzqdzzmv8l",
		"m/44'/988'/1'": "COUNTDOWN13ss78hz4epq88putspwmf5ks9m7g8cv2wdqqut",
		"m/44'/988'/2'": "COUNTDOWN160jdxeaxyfjrqcfmf5cdzpujyz25nphtpm63rp",
		"m/44'/988'/3'": "COUNTDOWN1x3wpuhvrm5vvex38k43csxdjs7l4n3gwjpt08w",
		"m/44'/988'/4'": "COUNTDOWN1fm8wddz5mk8xeqcesleysrfd5pg9wg0xh5wj2v",
	}

	for path, bech := range cases {
		t.Run(path, func(t *testing.T) {
			priv, err := keygen(mnemonic, path)
			if err != nil {
				t.Fatalf("cannot generate key: %s", err)
			}
			b, err := toBech32("COUNTDOWN", priv.Public().(ed25519.PublicKey))
			if err != nil {
				t.Fatalf("cannot serialize to bech32: %s", err)
			}
			if got := string(b); got != bech {
				t.Logf("want: %s", bech)
				t.Logf(" got: %s", got)
				t.Fatal("unexpected bech address")
			}
		})
	}
}

func TestMnemonic(t *testing.T) {
	cases := map[string]struct {
		mnemonic string
		wantErr  bool
	}{

		"valid mnemonic 12 words": {
			mnemonic: "super bulk plunge better rookie donor reward obscure rescue type trade pelican",
			wantErr:  false,
		},
		"valid mnemonic 15 words": {
			mnemonic: "say debris entry orange grief deer train until flock scrub volume artist skill obscure immense",
			wantErr:  false,
		},
		"valid mnemonic 18 words": {
			mnemonic: "fetch height snow poverty space follow seven festival wasp pet asset tattoo cement twist exile trend bench eternal",
			wantErr:  false,
		},
		"valid mnemonic 21 words": {
			mnemonic: "increase shine pumpkin curtain trash cabbage juice canal ugly naive name insane indoor assault snap taxi casual unhappy buddy defense artefact",
			wantErr:  false,
		},
		"valid mnemonic 24 words": {
			mnemonic: "usage mountain noodle inspire distance lyrics caution wait mansion never announce biology squirrel guess key gain belt same matrix chase mom beyond model toy",
			wantErr:  false,
		},
		"additional whitespace around mnemonnic is not allowed (beginning)": {
			mnemonic: " super bulk plunge better rookie donor reward obscure rescue type trade pelican",
			wantErr:  true,
		},
		"additional whitespace around mnemonnic is not allowed (end)": {
			mnemonic: "super bulk plunge better rookie donor reward obscure rescue type trade pelican ",
			wantErr:  true,
		},
		"additional whitespace around mnemonnic is not allowed (middle)": {
			mnemonic: "super bulk plunge better rookie    donor reward obscure rescue type trade pelican",
			wantErr:  true,
		},
		"mnemonnic cannot be tab separated": {
			mnemonic: "super\tbulk plunge better rookie donor reward obscure rescue type trade pelican",
			wantErr:  true,
		},
		"mnenomic that is valid in a language other than English (Italian)": {
			mnemonic: "acrobata acuto adagio addebito addome adeguato aderire adipe adottare adulare affabile affetto affisso affranto aforisma",
			wantErr:  true,
		},
		"mnenomic that is valid in a language other than English (Japanese)": {
			mnemonic: " あつかう あっしゅく あつまり あつめる あてな あてはまる あひる あぶら あぶる あふれる あまい あまど ",
			wantErr:  true,
		},
		"initially valid mnemonic that the last word was changed": {
			mnemonic: "super bulk plunge better rookie donor reward obscure rescue type trade trade",
			wantErr:  true,
		},
	}

	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			_, err := keygen(tc.mnemonic, "m/44'/988'/0'")
			if hasErr := err != nil; hasErr != tc.wantErr {
				t.Fatalf("returned erorr value: %+v", err)
			}
		})
	}
}
