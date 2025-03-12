using System;
using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace TYBLOG.Migrations
{
    /// <inheritdoc />
    public partial class AddReviewedStatus : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.CreateTable(
                name: "Applications",
                columns: table => new
                {
                    Id = table.Column<int>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    DiscordUsername = table.Column<string>(type: "TEXT", nullable: false),
                    MinecraftUsername = table.Column<string>(type: "TEXT", nullable: false),
                    Age = table.Column<int>(type: "INTEGER", nullable: false),
                    FavoriteAboutMinecraft = table.Column<string>(type: "TEXT", nullable: false),
                    UnderstandingOfSMP = table.Column<string>(type: "TEXT", nullable: false),
                    JoinedDiscord = table.Column<bool>(type: "INTEGER", nullable: false),
                    SubmissionDate = table.Column<DateTime>(type: "TEXT", nullable: false),
                    IsReviewed = table.Column<bool>(type: "INTEGER", nullable: false),
                    ReviewedAt = table.Column<DateTime>(type: "TEXT", nullable: true),
                    ReviewNotes = table.Column<string>(type: "TEXT", nullable: true)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Applications", x => x.Id);
                });

            migrationBuilder.CreateIndex(
                name: "IX_Applications_MinecraftUsername",
                table: "Applications",
                column: "MinecraftUsername",
                unique: true);
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropTable(
                name: "Applications");
        }
    }
}
