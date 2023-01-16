package main

import (
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	"strings"

	"facette.io/natsort"
	"github.com/cheggaaa/pb/v3"
	"github.com/jung-kurt/gofpdf"
	"golang.org/x/text/unicode/norm"

	_ "image/jpeg"
	_ "image/png"
)

func FindFile(root string) [][]string {
	dirEntries, err := os.ReadDir(root)
	if err != nil {
		log.Fatal(err)
	}
	// 直下に画像があるかどうかだけ先に調べる
	hasPic := false
	for _, e := range dirEntries {
		if !e.IsDir() { // ファイルのみ
			ext := strings.ToLower(filepath.Ext(e.Name()))
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" { // とりあえずjpgとpngのみ
				hasPic = true
			}
		}
	}
	var out [][]string
	if hasPic { // 画像があった
		buff := make([]string, 0, 1000)
		for _, e := range dirEntries {
			if !e.IsDir() {
				if ext := strings.ToLower(filepath.Ext(e.Name())); ext == ".jpg" || ext == ".jpeg" || ext == ".png" { // とりあえずjpgとpngのみ
					if len(buff) == 0 {
						buff = append(buff, root) // 0番目にディレクトリ名を入れる
					}
					buff = append(buff, filepath.Join(root, e.Name()))
				}
			}
		}
		natsort.Sort(buff)
		out = append(out, buff)
	} else { // 画像がなかった 1階層だけ深く探索する それ以上にはいかない
		for _, e := range dirEntries {
			if e.IsDir() {
				dir := filepath.Join(root, e.Name())
				dirEntries, err := os.ReadDir(dir)
				if err != nil {
					log.Fatal(err)
				}
				buff := make([]string, 0, 1000)
				for _, e := range dirEntries { // 拡張子が画像のファイルパスを得る
					if !e.IsDir() {
						if ext := strings.ToLower(filepath.Ext(e.Name())); ext == ".jpg" || ext == ".jpeg" || ext == ".png" { // とりあえずjpgとpngのみ
							if len(buff) == 0 {
								buff = append(buff, dir) // 0番目にディレクトリ名を入れる
							}
							buff = append(buff, filepath.Join(dir, e.Name()))
						}
					}
				}
				natsort.Sort(buff)
				if len(buff) != 0 {
					out = append(out, buff)
				}
			}
		}
	}
	return out
}

func getImgSize(imgPath string) []int {
	file, err := os.Open(imgPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	return []int{img.Bounds().Dx(), img.Bounds().Dy()}
}

func main() {
	log.SetFlags(log.Lshortfile | log.Ltime | log.Lmicroseconds)
	//log.Printf("cpu: %d\n", runtime.NumCPU())

	// 第一引数に画像ファイルを含むフォルダのパスをもらう
	// 直下に画像ファイルがある場合それらをまとめてpdfにする。それ以下の改装の探索はしない。
	// もし直下ではなく2階層下にのみ画像がある場合は2階層下まで探索する
	v := FindFile(os.Args[1])
	// これでvに画像のパスのリストが入る。ただし各リストの0番目にはフォルダ名が入っている

	for _, vv := range v { // 画像があるディレクトリごとにリストになっている
		name := norm.NFC.String(filepath.Base(vv[0])) + ".pdf" // 画像があるディレクトリ名が入っている Unicode正規化
		pdf := gofpdf.New("P", "in", "A4", "")                 // Pはポートレート(Lにするとランドスケープ) inはインチ 末尾の空文字はフォント指定
		// 1ページごとの処理
		// 画像のサイズを調べ72dpi相当でインチに変換する
		// 画像サイズ(pixel) からmmへの変換がインチを挟むので仕方なくインチで用紙サイズを定義する
		// ヤードポンド法は滅びろ

		fmt.Println("Generating pdf file : " + name)
		// 進捗表示
		bar := pb.Simple.Start(len(vv) - 1)
		bar.SetMaxWidth(80)
		for _, imgPath := range vv[1:] {
			size := getImgSize(imgPath)                                               // これはpixel単位
			sizeInch := []float64{float64(size[0]) / 72., float64(size[1]) / 72.}     // 忌々しいインチ
			pdf.AddPageFormat("P", gofpdf.SizeType{Wd: sizeInch[0], Ht: sizeInch[1]}) // 画像のサイズに合わせる
			pdf.Image(imgPath, 0, 0, sizeInch[0], sizeInch[1], false, "", 0, "")
			bar.Increment()
		}
		bar.Finish()
		// pdfの保存先は第1引数の親ディレクトリ
		output := filepath.Join(filepath.Dir(os.Args[1]), name)
		fmt.Println("Wrighting pdf file to " + output)
		if err := pdf.OutputFileAndClose(output); err == nil {
			fmt.Println("PDF " + name + " generated successfully")
		}
	}
}
