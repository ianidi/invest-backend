package mail

import (
	"fmt"
	"html/template"
	"io"
	"net/smtp"
	"os"
	"time"

	"github.com/domodwyer/mailyak"
	"github.com/ianidi/exchange-server/internal/db"
	"github.com/ianidi/exchange-server/internal/models"
	"github.com/spf13/cast"
)

//SendMail отправить письмо по электронной почте
func SendMail(email string, subject string, title string, content string, button string, link string, description string) {
	db := db.GetDB()

	var settings models.Settings

	err := db.Get(&settings, "SELECT * FROM Settings WHERE SettingsID=$1", 1)

	if err != nil {
		fmt.Print(err)
		return
	}

	// Create a new email - specify the SMTP host and auth
	mail := mailyak.New(settings.SMTPHost+":"+cast.ToString(settings.SMTPPort), smtp.PlainAuth("", settings.SMTPUsername, settings.SMTPPassword, settings.SMTPHost))

	mail.To(email)
	mail.From(settings.SMTPFromEmail)
	mail.FromName(settings.SMTPFromName)

	mail.Subject(subject)

	template, err := template.New("htmlEmail").Parse(`{{define "htmlEmail"}}<!DOCTYPE html>
	<html lang="%1$s">
		<head>
			<meta charset="UTF-8" />
			<meta http-equiv="Content-Type" content="text/html;charset=utf8" />
			<title>%2$s</title>
		</head>
		<link href="https://fonts.googleapis.com/css?family=Roboto:300,400&amp;subset=cyrillic-ext" rel="stylesheet" />
		<!--[if mso]>
			<link href="https://fonts.googleapis.com/css?family=Roboto:300,400&amp;subset=cyrillic-ext" rel="stylesheet" />
		<![endif]-->
		<div
			style="
				mso-hide: all;
				visibility: hidden;
				opacity: 0;
				color: transparent;
				font-size: 0px;
				width: 0;
				height: 0;
				display: none !important;
			"
			class="preheader"
		></div>
		<table
			style="
				border-spacing: 0;
				border-collapse: collapse;
				vertical-align: top;
				-webkit-hyphens: none;
				-moz-hyphens: none;
				hyphens: none;
				-ms-hyphens: none;
				font-family: 'Roboto', Helvetica, Arial, sans-serif;
				font-weight: normal;
				margin: 0;
				text-align: left;
				font-size: 16px;
				line-height: 19px;
				background: #f3f3f3;
				padding: 0;
				width: 100%;
				height: 100%;
				color: #0a0a0a;
				margin-bottom: 0px !important;
				background-color: white;
			"
			class="body"
		>
			<tbody>
				<tr style="padding: 0; vertical-align: top; text-align: left;">
					<td
						style="
							font-size: 16px;
							word-wrap: break-word;
							-webkit-hyphens: auto;
							-moz-hyphens: auto;
							hyphens: auto;
							vertical-align: top;
							text-align: left;
							line-height: 1.3;
							color: #0a0a0a;
							font-family: 'Roboto', Helvetica, Arial, sans-serif;
							padding: 0;
							margin: 0;
							font-weight: normal;
							border-collapse: collapse !important;
						"
						valign="top"
					>
						<center style="width: 100%; min-width: 580px;">
							<table
								style="
									border-spacing: 0;
									border-collapse: collapse;
									padding: 0;
									vertical-align: top;
									background: #fefefe;
									width: 580px;
									margin: 0 auto;
									text-align: inherit;
									max-width: 580px;
								"
								class="container"
							>
								<tbody>
									<tr style="padding: 0; vertical-align: top; text-align: left;">
										<td
											style="
												font-size: 16px;
												word-wrap: break-word;
												-webkit-hyphens: auto;
												-moz-hyphens: auto;
												hyphens: auto;
												vertical-align: top;
												text-align: left;
												line-height: 1.3;
												color: #0a0a0a;
												font-family: 'Roboto', Helvetica, Arial, sans-serif;
												padding: 0;
												margin: 0;
												font-weight: normal;
												border-collapse: collapse !important;
											"
										>
											<div>
												<table
													style="
														border-spacing: 0;
														border-collapse: collapse;
														text-align: left;
														vertical-align: top;
														padding: 0;
														width: 100%;
														position: relative;
														display: table;
													"
													class="row"
												>
													<tbody>
														<tr style="padding: 0; vertical-align: top; text-align: left;" class="">
															<th
																style="
																	font-size: 16px;
																	padding: 0;
																	text-align: left;
																	color: #0a0a0a;
																	font-family: 'Roboto', Helvetica, Arial, sans-serif;
																	font-weight: normal;
																	line-height: 1.3;
																	margin: 0 auto;
																	padding-bottom: 16px;
																	width: 564px;
																	padding-left: 16px;
																	padding-right: 16px;
																"
																class="columns first large-12 last small-12"
															>
																<div
																	style="
																		font-family: 'Roboto', Helvetica, Arial, sans-serif;
																		font-weight: normal;
																		padding: 0;
																		margin: 0;
																		text-align: left;
																		line-height: 1.3;
																		color: #2199e8;
																		text-decoration: none;
																	"
																>
																	<img
																		style="
																			display: block;
																			outline: none;
																			text-decoration: none;
																			-ms-interpolation-mode: bicubic;
																			width: auto;
																			max-width: 100%;
																			clear: both;
																			border: none;
																			padding-top: 48px;
																			padding-bottom: 16px;
																			max-height: 40px;
																		"
																		src="cid:logo.png"
																		width="60"
																		height="37"
																		alt=""
																	/>
																</div>
															</th>
														</tr>
													</tbody>
												</table>
											</div>
											<div>
												<table
													style="
														border-spacing: 0;
														border-collapse: collapse;
														text-align: left;
														vertical-align: top;
														padding: 0;
														width: 100%;
														position: relative;
														display: table;
													"
													class="row"
												>
													<tbody>
														<tr style="padding: 0; vertical-align: top; text-align: left;" class="">
															<th
																style="
																	font-size: 16px;
																	padding: 0;
																	text-align: left;
																	color: #0a0a0a;
																	font-family: 'Roboto', Helvetica, Arial, sans-serif;
																	font-weight: normal;
																	line-height: 1.3;
																	margin: 0 auto;
																	padding-bottom: 16px;
																	width: 564px;
																	padding-left: 16px;
																	padding-right: 16px;
																"
																class="columns first large-12 last small-12"
															>
																<p
																	style="
																		font-family: 'Roboto', Helvetica, Arial, sans-serif;
																		padding: 0;
																		margin: 0;
																		line-height: 1.4;
																		font-weight: 400;
																		color: #484848;
																		font-size: 20px;
																		hyphens: none;
																		-ms-hyphens: none;
																		-webkit-hyphens: none;
																		-moz-hyphens: none;
																		text-align: left;
																		margin-bottom: 0px !important;
																	"
																	class="body body-lg body-link-rausch light text-left"
																>
																	{{.title}}
																</p>
															</th>
														</tr>
													</tbody>
												</table>
											</div>
											<div>
												<table
													style="
														border-spacing: 0;
														border-collapse: collapse;
														text-align: left;
														vertical-align: top;
														padding: 0;
														width: 100%;
														position: relative;
														display: table;
													"
													class="row"
												>
													<tbody>
														<tr style="padding: 0; vertical-align: top; text-align: left;" class="">
															<th
																style="
																	font-size: 16px;
																	padding: 0;
																	text-align: left;
																	color: #0a0a0a;
																	font-family: 'Roboto', Helvetica, Arial, sans-serif;
																	font-weight: normal;
																	line-height: 1.3;
																	margin: 0 auto;
																	padding-bottom: 16px;
																	width: 564px;
																	padding-left: 16px;
																	padding-right: 16px;
																"
																class="columns first large-12 last small-12"
															>
																<p
																	style="
																		font-family: 'Roboto', Helvetica, Arial, sans-serif;
																		padding: 0;
																		margin: 0;
																		line-height: 1.4;
																		font-weight: 300;
																		color: #484848;
																		font-size: 18px;
																		hyphens: none;
																		-ms-hyphens: none;
																		-webkit-hyphens: none;
																		-moz-hyphens: none;
																		text-align: left;
																		margin-bottom: 0px !important;
																	"
																	class="body body-lg body-link-rausch light text-left"
																>
																	{{.content}}
																</p>
															</th>
														</tr>
													</tbody>
												</table>
											</div>
											<div style="padding-top: 8px;">
												<table
													style="
														border-spacing: 0;
														border-collapse: collapse;
														text-align: left;
														vertical-align: top;
														padding: 0;
														width: 100%;
														position: relative;
														display: table;
													"
													class="row"
												>
													<tbody>
														<tr style="padding: 0; vertical-align: top; text-align: left;">
															<th
																style="
																	color: #0a0a0a;
																	font-family: 'Roboto', Helvetica, Arial, sans-serif;
																	font-weight: normal;
																	padding: 0;
																	margin: 0;
																	text-align: left;
																	font-size: 16px;
																	line-height: 1.3;
																	padding-left: 16px;
																	padding-right: 16px;
																"
																class="col-pad-left-2 col-pad-right-2"
															>
																<a
																	style="
																		font-family: 'Roboto', Helvetica, Arial, sans-serif;
																		background-color: #2fa8ec;
																		border-color: #2fa8ec;
																		border-style: solid;
																		border-width: 13px 16px;
																		color: #ffffff;
																		display: inline-block;
																		max-width: 300px;
																		min-width: 150px;
																		-webkit-border-radius: 3px;
																		border-radius: 3px;
																		text-align: center;
																		text-decoration: none;
																		transition: all 0.2s ease-in;
																	"
																	href="{{.link}}"
																>
																	<span style="float: left; text-align: left;">{{.button}}</span> <span style="float:right;padding-top:2px; display:inline-block;"> <img id="OWATemporaryImageDivContainer1621827" class="" alt="" style="-ms-interpolation-mode: bicubic; Margin-left: 16px; border: none; clear: both; display: block; margin-top: 2px; max-width: 100%; outline: none; text-decoration: none; width: auto;" height="12" width="16" src="cid:arrow.png"></span></a
																>
															</th>
														</tr>
													</tbody>
												</table>
											</div>
											<div style="padding-top: 24px;">
												<table
													style="
														border-spacing: 0;
														border-collapse: collapse;
														text-align: left;
														vertical-align: top;
														padding: 0;
														width: 100%;
														position: relative;
														display: table;
													"
													class="row"
												>
													<tbody>
														<tr style="padding: 0; vertical-align: top; text-align: left;" class="">
															<th
																style="
																	font-size: 16px;
																	padding: 0;
																	text-align: left;
																	color: #0a0a0a;
																	font-family: 'Roboto', Helvetica, Arial, sans-serif;
																	font-weight: normal;
																	line-height: 1.3;
																	margin: 0 auto;
																	padding-bottom: 16px;
																	width: 564px;
																	padding-left: 16px;
																	padding-right: 16px;
																"
																class="columns first large-12 last small-12"
															>
																<p
																	style="
																		font-family: 'Roboto', Helvetica, Arial, sans-serif;
																		padding: 0;
																		margin: 0;
																		line-height: 1.4;
																		font-weight: 300;
																		color: #484848;
																		font-size: 18px;
																		hyphens: none;
																		-ms-hyphens: none;
																		-webkit-hyphens: none;
																		-moz-hyphens: none;
																		text-align: left;
																		margin-bottom: 0px !important;
																	"
																	class="body body-lg body-link-rausch light text-left"
																>
																	{{.description}}
																</p>
															</th>
														</tr>
													</tbody>
												</table>
											</div>
	
											<div style="padding-top: 30px;">
												<table
													style="
														border-spacing: 0;
														border-collapse: collapse;
														text-align: left;
														vertical-align: top;
														padding: 0;
														width: 100%;
														position: relative;
														display: table;
													"
													class="row"
												>
													<tbody>
														<tr style="padding: 0; vertical-align: top; text-align: left;" class="">
															<th
																style="
																	font-size: 16px;
																	padding: 0;
																	text-align: left;
																	color: #0a0a0a;
																	font-family: 'Roboto', Helvetica, Arial, sans-serif;
																	font-weight: normal;
																	line-height: 1.3;
																	margin: 0 auto;
																	padding-bottom: 16px;
																	width: 564px;
																	padding-left: 16px;
																	padding-right: 16px;
																"
																class="columns first large-12 last small-12"
															>
																<p
																	style="
																		font-family: 'Roboto', Helvetica, Arial, sans-serif;
																		padding: 0;
																		margin: 0;
																		line-height: 1.4;
																		font-weight: 300;
																		color: #8c94a0;
																		font-size: 16px;
																		hyphens: none;
																		-ms-hyphens: none;
																		-webkit-hyphens: none;
																		-moz-hyphens: none;
																		text-align: left;
																		margin-bottom: 0px !important;
																	"
																	class="body body-lg body-link-rausch light text-left"
																>
																	{{.footer}}
																</p>
															</th>
														</tr>
													</tbody>
												</table>
											</div>
										</td>
									</tr>
								</tbody>
							</table>
						</center>
					</td>
				</tr>
			</tbody>
		</table>
	</html>
	{{end}}`)

	if err != nil {
		fmt.Print(err)
		return
	}

	logo, err := os.Open("./email/logo.png")
	defer logo.Close()

	if err != nil {
		fmt.Print(err)
		return
	}

	arrow, err := os.Open("./email/arrow.png")
	defer arrow.Close()

	if err != nil {
		fmt.Print(err)
		return
	}

	var logoReader io.Reader
	logoReader = logo

	var arrowReader io.Reader
	arrowReader = arrow
	mail.Attach("logo.png", logoReader)
	mail.Attach("arrow.png", arrowReader)

	//"If you have further questions, please feel free to use the live chat."
	if err := template.ExecuteTemplate(mail.HTML(), "htmlEmail", map[string]string{"title": title, "content": content, "button": button, "link": link, "description": description, "footer": "©" + cast.ToString(time.Now().Year()) + " " + settings.Title}); err != nil {
		fmt.Print(err)
		return
	}

	//mail.Plain().Set(content)

	if err := mail.Send(); err != nil {
		fmt.Print(err)
		return
	}
}
