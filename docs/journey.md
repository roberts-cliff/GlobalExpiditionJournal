2. Detailed User Journey & App Flow
   Primary Navigation & Core Pages
   The app is structured around a main menu, accessible from most screens, leading to six core sections:
   The Passport (Your Travel Hub)
   The Scrapbook (Your Memories)
   The Library (Your Knowledge Base)
   The Checklist (Your Travel To-Do's)
   The Awards Page (Your Achievements)
   The Wishlist (Your Travel Dreams)
   Here is a visual overview of how these core pages connect, followed by detailed interaction flows:
```mermaid
   flowchart TD
   A[App Opens<br>Opening Pages & Squares] --> B[Main Menu / Home]

   B --> P[Passport Page]
   B --> S[Scrapbook]
   B --> L[Library]
   B --> C[Checklist]
   B --> W[Wishlist]
   B --> AW[Awards Page]

   P --> P_Map[Map View<br>All Countries]
   P --> P_List[List View<br>All Countries]
   P --> P_Stats[Your Stats]

   P_Map --> P_C[Specific Country Page]
   P_List --> P_C

   S --> S_C[Country Scrapbook<br>Media & Notes]
   S_C --> S_Add[Add Photo/Note]

   C --> C_Item[Checklist Item<br>e.g., 'Try Local Food']
   C_Item --> C_Check[Mark Complete]
   C_Check --> AW_Update[Awards Page Updates]

   W --> W_Add[Add New Destination]

   AW --> AW_Detail[Award Detail View]
```

Interaction Flow: "Given → When → Then" Examples
Journey 1: Documenting a Visit to a Country
GIVEN a user is viewing their Passport's map view.
WHEN they tap on a country they have visited (e.g., France).
THEN they are taken to the Country Detail Page (from "Passport Page Purple" mockup). Here they can:
Log the date of visit.
See a summary of activities/stamps collected.
Tap a button like "Go to Scrapbook".
WHEN they tap "Go to Scrapbook".
THEN they navigate to the "Country Scrapbook Blue" page for that nation. Here they can:
View a gallery of past entries.
Tap an "Add+" button to create a new memory.
WHEN they tap "Add+".
THEN a modal or new screen opens, allowing them to add a photo, write notes, tag a location, and link this memory to a specific "stamp" or activity from the checklist (e.g., "Tried Croissant").
Journey 2: Completing a Travel Activity (Checklist to Awards)
GIVEN a user is on a trip and opens the "Checklist Page White".
WHEN they find an activity (e.g., "Visit a World Heritage Site" under the "Cultural" category) and tap it.
THEN the checklist item expands (as shown in mockups) to show details, notes, or an option to mark it complete.
WHEN they tap to mark it complete.
THEN:
The item is checked off their list.
A corresponding "stamp" is added to that country's page in their Passport.
Their progress bar(s) on the Awards Page (from "Awards and Completion Page Red") update automatically.
IF completing this item fulfills a collection (e.g., all "Cultural" items for France), THEN a new Award/Badge is unlocked and displayed prominently on the Awards Page.
Journey 3: Planning a Future Trip
GIVEN a user is dreaming of a future destination.
WHEN they go to the "Wishlist Page Yellow" and tap "Add a New Destination".
THEN they can search for and select a country.
WHEN a country is added to their Wishlist.
THEN:
It appears as a "planned" or "wishlisted" flag on their Passport map.
They can go to that country's page and pre-populate a future Scrapbook with links, plans, and a pre-trip checklist.
The Library section can be used to save relevant articles about that destination for later review.
Key Connections Between Features (How Everything Links)
Passport ↔ Scrapbook: The Passport is the index; every country links to its dedicated Scrapbook for memories.
Checklist ↔ Awards ↔ Passport: Completing checklist items grants stamps (in Passport) and progress toward awards.
Wishlist ↔ Passport ↔ Checklist: A wishlisted country can have its own pre-trip checklist and will appear visually distinct in the Passport.
Library ↔ All: Saved research in the Library can be linked to specific countries in the Passport, Scrapbook, or Wishlist.

