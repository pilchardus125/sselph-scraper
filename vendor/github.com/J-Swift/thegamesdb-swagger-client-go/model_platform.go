/*
 * TheGamesDB API
 *
 * API Documentation
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package gamesdb

// Platform struct for Platform
type Platform struct {
	Id         int32  `json:"id"`
	Name       string `json:"name"`
	Alias      string `json:"alias"`
	Icon       string `json:"icon"`
	Console    string `json:"console"`
	Controller string `json:"controller"`
	Developer  string `json:"developer"`
	Overview   string `json:"overview,omitempty"`
}