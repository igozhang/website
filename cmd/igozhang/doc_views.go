package main

import "net/http"

const docViewCookieMaxAge = 12 * 60 * 60 // 12 hours

func docViewCookieName(slug string) string {
	return "igozhang_doc_v_" + slug
}

func docViewLegacyCookieName(slug string) string {
	return "igo_doc_v_" + slug
}

func docViewCookieSeen(r *http.Request, slug string) bool {
	for _, name := range []string{docViewCookieName(slug), docViewLegacyCookieName(slug)} {
		c, err := r.Cookie(name)
		if err == nil && c.Value == "1" {
			return true
		}
	}
	return false
}

func setDocViewCookie(w http.ResponseWriter, slug string) {
	http.SetCookie(w, &http.Cookie{
		Name:     docViewCookieName(slug),
		Value:    "1",
		Path:     "/",
		MaxAge:   docViewCookieMaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// recordDocView returns current view count; increments at most once per browser per slug per 12h.
func recordDocView(w http.ResponseWriter, r *http.Request, slug string) (int64, error) {
	meta, err := docStore.Get(slug)
	if err != nil {
		return 0, err
	}
	if r.Method == http.MethodHead {
		return meta.Views, nil
	}
	if docViewCookieSeen(r, slug) {
		return meta.Views, nil
	}
	setDocViewCookie(w, slug)
	if err := docStore.IncrViews(slug); err != nil {
		return meta.Views, err
	}
	return meta.Views + 1, nil
}
